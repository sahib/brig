package db

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	capnp "github.com/sahib/brig/gateway/db/capnp"
	"github.com/sahib/brig/util"
	log "github.com/sirupsen/logrus"
	capnp_lib "zombiezen.com/go/capnproto2"
)

const (
	// RightDownload refers to the right to download/view a file.
	RightDownload = "fs.download"
	// RightFsView refers to the right to view everything related to
	// the filesystem and history.
	RightFsView = "fs.view"
	// RightFsEdit refers to the right to edit the filesystem.
	// This includes pinning.
	RightFsEdit = "fs.edit"
	// RightRemotesView is the right to view the remote list.
	RightRemotesView = "remotes.view"
	// RightRemotesEdit is the right to edit the remote list.
	RightRemotesEdit = "remotes.edit"
)

var (
	// DefaultRights is a list of rights that users will get
	// if no other explicit rights are given. They are identical
	// to the admin role currently.
	DefaultRights = []string{
		RightDownload,
		RightFsView,
		RightFsEdit,
		RightRemotesView,
		RightRemotesEdit,
	}

	// AllRights is a map that can be quickly used to check
	// if a right is valid or not.
	AllRights = map[string]bool{
		RightDownload:    true,
		RightFsView:      true,
		RightFsEdit:      true,
		RightRemotesView: true,
		RightRemotesEdit: true,
	}
)

// UserDatabase is a badger db that stores user information,
// using the user name as unique key.
type UserDatabase struct {
	isStopped int64
	mu        sync.Mutex
	db        *badger.DB
	gcTicker  *time.Ticker
}

// NewUserDatabase creates a new UserDatabase at `path` or loads
// an existing one.
func NewUserDatabase(path string) (*UserDatabase, error) {
	opts := badger.DefaultOptions(path).
		WithValueLogFileSize(10 * 1024 * 1024). //default is 2GB we should not need 2GB
		WithMemTableSize(10 * 1024 * 1024).     //default is 64MB
		WithSyncWrites(false).
		WithLogger(nil)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	gcTicker := time.NewTicker(5 * time.Minute)

	udb := &UserDatabase{db: db, gcTicker: gcTicker}

	go func() {
		for range gcTicker.C {
			if atomic.LoadInt64(&udb.isStopped) > 0 {
				return
			}
			if err := db.RunValueLogGC(0.5); err != nil {
				if err != badger.ErrNoRewrite {
					log.WithError(err).Warnf("badger gc failed")
				}
			}

		}
	}()

	return udb, nil
}

// Close cleans up all the resources used by a badger db.
func (ub *UserDatabase) Close() error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	ub.gcTicker.Stop()
	atomic.StoreInt64(&ub.isStopped, 1)

	if err := ub.db.Close(); err != nil {
		return err
	}

	ub.db = nil
	return nil
}

func unmarshalUser(data []byte) (*User, error) {
	msg, err := capnp_lib.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	capUser, err := capnp.ReadRootUser(msg)
	if err != nil {
		return nil, err
	}

	return UserFromCapnp(capUser)
}

// UserFromCapnp takes a capnp.user and returns a regular User from it.
func UserFromCapnp(capUser capnp.User) (*User, error) {
	capFolders, err := capUser.Folders()
	if err != nil {
		return nil, err
	}

	folders := []string{}
	for idx := 0; idx < capFolders.Len(); idx++ {
		folder, err := capFolders.At(idx)
		if err != nil {
			return nil, err
		}

		folders = append(folders, folder)
	}

	capRights, err := capUser.Rights()
	if err != nil {
		return nil, err
	}

	rights := []string{}
	for idx := 0; idx < capRights.Len(); idx++ {
		right, err := capRights.At(idx)
		if err != nil {
			return nil, err
		}

		rights = append(rights, right)
	}

	name, err := capUser.Name()
	if err != nil {
		return nil, err
	}

	passwordHash, err := capUser.PasswordHash()
	if err != nil {
		return nil, err
	}

	salt, err := capUser.Salt()
	if err != nil {
		return nil, err
	}

	return &User{
		Name:         name,
		PasswordHash: passwordHash,
		Salt:         salt,
		Folders:      folders,
		Rights:       rights,
	}, nil
}

func marshalUser(user *User) ([]byte, error) {
	msg, seg, err := capnp_lib.NewMessage(capnp_lib.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	if _, err := UserToCapnp(user, seg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

// UserToCapnp converts a User to a capnp.User.
func UserToCapnp(user *User, seg *capnp_lib.Segment) (*capnp.User, error) {
	capUser, err := capnp.NewRootUser(seg)
	if err != nil {
		return nil, err
	}

	capFolders, err := capnp_lib.NewTextList(seg, int32(len(user.Folders)))
	if err != nil {
		return nil, err
	}

	for idx, folder := range user.Folders {
		if err := capFolders.Set(idx, folder); err != nil {
			return nil, err
		}
	}

	if err := capUser.SetFolders(capFolders); err != nil {
		return nil, err
	}

	capRights, err := capnp_lib.NewTextList(seg, int32(len(user.Rights)))
	if err != nil {
		return nil, err
	}

	for idx, right := range user.Rights {
		if err := capRights.Set(idx, right); err != nil {
			return nil, err
		}
	}

	if err := capUser.SetRights(capRights); err != nil {
		return nil, err
	}

	if err := capUser.SetName(user.Name); err != nil {
		return nil, err
	}

	if err := capUser.SetPasswordHash(user.PasswordHash); err != nil {
		return nil, err
	}

	if err := capUser.SetSalt(user.Salt); err != nil {
		return nil, err
	}

	return &capUser, nil
}

// User is one user that is stored in the database.
// The passwords are stored as scrypt hash with added salt.
type User struct {
	Name         string
	PasswordHash string
	Salt         string
	Folders      []string
	Rights       []string
}

// CheckPassword checks if `password` matches the stored one.
func (u User) CheckPassword(password string) (bool, error) {
	salt, err := base64.StdEncoding.DecodeString(u.Salt)
	if err != nil {
		return false, err
	}

	oldHash, err := base64.StdEncoding.DecodeString(u.PasswordHash)
	if err != nil {
		return false, err
	}

	newHash := util.DeriveKey([]byte(password), salt, 32)
	return subtle.ConstantTimeCompare(oldHash, newHash) == 1, nil
}

// HashPassword creates a new hash and salt from a password.
func HashPassword(password string) (string, string, error) {
	// Read a new salt from a random source.
	// 8 bytes are considered enough by the scrypt documentation.
	salt := make([]byte, 8)
	if n, err := rand.Read(salt); err != nil {
		return "", "", err
	} else if n != 8 {
		return "", "", fmt.Errorf("did not read enough random bytes")
	}

	// Derive the actual hash and encode it to base64.
	hash := util.DeriveKey([]byte(password), salt, 32)
	encode := base64.StdEncoding.EncodeToString
	return encode(hash), encode(salt), nil
}

// Add adds a new user to the database.
// If the user exists already, it is overwritten.
func (ub *UserDatabase) Add(name, password string, folders []string, rights []string) error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	if len(folders) == 0 {
		folders = []string{"/"}
	}

	if len(rights) == 0 {
		rights = DefaultRights
	}

	for _, right := range rights {
		if !AllRights[right] {
			return fmt.Errorf("invalid right: %s", right)
		}
	}

	hashed, salt, err := HashPassword(password)
	if err != nil {
		return err
	}

	user := &User{
		Name:         name,
		PasswordHash: hashed,
		Salt:         salt,
		Folders:      folders,
		Rights:       rights,
	}

	data, err := marshalUser(user)
	if err != nil {
		return err
	}

	return ub.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(name), data)
	})
}

// Get returns a User, if it exists. If it does not exist,
// an error will be returned.
func (ub *UserDatabase) Get(name string) (User, error) {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	user := User{}
	return user, ub.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}

		return item.Value(func(data []byte) error {
			decUser, err := unmarshalUser(data)
			if err != nil {
				return err
			}

			user = *decUser
			return nil
		})
	})
}

// Remove removes an existing user.
func (ub *UserDatabase) Remove(name string) error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	return ub.db.Update(func(txn *badger.Txn) error {
		// Make sure to error out if the key did not exist:
		if _, err := txn.Get([]byte(name)); err != nil {
			return err
		}

		return txn.Delete([]byte(name))
	})
}

// List returns all users currently in the database.
func (ub *UserDatabase) List() ([]User, error) {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	users := []User{}
	return users, ub.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.IteratorOptions{})
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(data []byte) error {
				user, err := unmarshalUser(data)
				if err != nil {
					return err
				}

				users = append(users, *user)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})
}
