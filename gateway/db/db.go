package db

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"gx/ipfs/QmZ7bFqkoHU2ARF68y9fSQVKcmhjYrTQgtCQ4i3chwZCgQ/badger"
	"sync"

	"github.com/sahib/brig/util"
)

// UserDatabase is a badger db that stores user information,
// using the user name as unique key.
type UserDatabase struct {
	mu sync.Mutex
	db *badger.DB
}

// NewUserDatabase creates a new UserDatabase at `path` or loads
// an existing one.
func NewUserDatabase(path string) (*UserDatabase, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &UserDatabase{db: db}, nil
}

// Close cleans up all the resources used by a badger db.
func (ub *UserDatabase) Close() error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	if err := ub.db.Close(); err != nil {
		return err
	}

	ub.db = nil
	return nil
}

// User is one user that is stored in the database.
// The passwords are stored as scrypt hash with added salt.
type User struct {
	Name         string
	PasswordHash string
	Salt         string
	Folders      []string
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
func (ub *UserDatabase) Add(name, password string, folders []string) error {
	ub.mu.Lock()
	defer ub.mu.Unlock()

	buf := &bytes.Buffer{}

	hashed, salt, err := HashPassword(password)
	if err != nil {
		return err
	}

	if folders == nil {
		folders = []string{"/"}
	}

	user := &User{
		Name:         name,
		PasswordHash: hashed,
		Salt:         salt,
		Folders:      folders,
	}

	if err := gob.NewEncoder(buf).Encode(user); err != nil {
		return err
	}

	return ub.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(name), buf.Bytes())
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
			return gob.NewDecoder(bytes.NewReader(data)).Decode(&user)
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
				user := User{}
				if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&user); err != nil {
					return err
				}

				users = append(users, user)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})
}
