package db

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	fileDBHash        = "hash"
	fileDBStores      = "stores"
	fileDBStorePrefix = "store_"
)

// A FileDB uses the filesystem for an extremely simple
// database.
type FileDB struct {
	Dir  string
	lock sync.RWMutex
}

func (f *FileDB) CreateUser(username, password string) (UserID, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	userDir := f.usernameDir(username)
	if err := os.Mkdir(userDir, 0700); err != nil {
		if os.IsExist(err) {
			return "", errors.New("create user: user already exists")
		}
		return "", errors.Wrap(err, "create user")
	}

	if err := f.setupNewUserFields(username, password); err != nil {
		os.RemoveAll(userDir)
		return "", errors.Wrap(err, "create user")
	}

	return UserID(username), nil
}

func (f *FileDB) setupNewUserFields(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := f.writeUserField(username, fileDBHash, hash); err != nil {
		return err
	}
	if err := f.writeUserField(username, fileDBStores, []byte("[]")); err != nil {
		return err
	}
	return nil
}

func (f *FileDB) Login(username, password string) (UserID, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	hash, err := f.readUserField(username, fileDBHash)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("check login: user does not exist")
		}
		return "", errors.Wrap(err, "check login")
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return "", errors.New("check login: password incorrect")
	}
	return UserID(username), nil
}

func (f *FileDB) usernameDir(username string) string {
	nameHash := sha256.Sum256([]byte(username))
	nameStr := base64.StdEncoding.EncodeToString(nameHash[:])
	return filepath.Join(f.Dir, nameStr)
}

func (f *FileDB) readUserField(username, field string) ([]byte, error) {
	path := filepath.Join(f.usernameDir(username), field)
	return ioutil.ReadFile(path)
}

func (f *FileDB) writeUserField(username, field string, data []byte) error {
	path := filepath.Join(f.usernameDir(username), field)
	return ioutil.WriteFile(path, data, 0700)
}
