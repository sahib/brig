package db

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"gx/ipfs/QmZ7bFqkoHU2ARF68y9fSQVKcmhjYrTQgtCQ4i3chwZCgQ/badger"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/util"
)

type UserDatabase struct {
	db *badger.DB
}

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

func (ub *UserDatabase) Close() error {
	return nil
}

type User struct {
	Name         string
	PasswordHash string
	Salt         string
	Folders      []string
}

func (u User) CheckPassword(password string) (bool, error) {
	log.Warningf("check password: %v", password)

	salt, err := base64.StdEncoding.DecodeString(u.Salt)
	if err != nil {
		return false, err
	}

	oldHash, err := base64.StdEncoding.DecodeString(u.PasswordHash)
	if err != nil {
		return false, err
	}

	newHash := util.DeriveKey([]byte(password), salt, 32)
	log.Warningf("compare %x %x", oldHash, newHash)
	return subtle.ConstantTimeCompare(oldHash, newHash) == 1, nil
}

func HashPassword(password string) (string, string, error) {
	// Read a new salt from a random source.
	// 8 bytes are considered enough by the scrypt documentation.
	log.Warningf("hash password: %v", password)
	salt := make([]byte, 8)
	if n, err := rand.Read(salt); err != nil {
		return "", "", err
	} else if n != 8 {
		return "", "", fmt.Errorf("did not read enough randon bytes")
	}

	// Derive the actual hash and encode it to base64.
	hash := util.DeriveKey([]byte(password), salt, 32)
	encode := base64.StdEncoding.EncodeToString
	log.Warningf("result: %s", encode(hash), encode(salt))
	return encode(hash), encode(salt), nil
}

func (ub *UserDatabase) Add(name, password string, folders []string) error {
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

func (ub *UserDatabase) Get(name string) (User, error) {
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

func (ub *UserDatabase) Remove(name string) error {
	return ub.db.Update(func(txn *badger.Txn) error {
		// Make sure to error out if the key did not exist:
		if _, err := txn.Get([]byte(name)); err != nil {
			return err
		}

		return txn.Delete([]byte(name))
	})
}

func (ub *UserDatabase) List() ([]User, error) {
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
