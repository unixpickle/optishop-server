package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/unixpickle/essentials"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	fileDBHash       = "hash"
	fileDBStores     = "stores"
	fileDBListPrefix = "store_"
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
	if err := f.encodeUserField(username, fileDBStores, []*StoreRecord{}); err != nil {
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

func (f *FileDB) Stores(user UserID) ([]*StoreRecord, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	var stores []*StoreRecord
	if err := f.decodeUserField(string(user), fileDBStores, &stores); err != nil {
		return nil, errors.Wrap(err, "get stores")
	}
	return stores, nil
}

func (f *FileDB) AddStore(user UserID, info *StoreInfo) (StoreID, error) {
	uid, err := randomUID()
	if err != nil {
		return "", errors.Wrap(err, "add store")
	}
	storeID := StoreID(uid)

	f.lock.Lock()
	defer f.lock.Unlock()

	var stores []*StoreRecord
	if err := f.decodeUserField(string(user), fileDBStores, &stores); err != nil {
		return "", errors.Wrap(err, "add store")
	}
	if err := f.encodeUserField(string(user), f.listField(storeID), []*ListEntry{}); err != nil {
		return "", errors.Wrap(err, "add store")
	}
	stores = append(stores, &StoreRecord{
		ID:   storeID,
		Info: info,
	})
	if err := f.encodeUserField(string(user), fileDBStores, &stores); err != nil {
		f.deleteUserField(string(user), f.listField(storeID))
		return "", errors.Wrap(err, "add store")
	}
	return storeID, nil
}

func (f *FileDB) RemoveStore(user UserID, store StoreID) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	var stores []*StoreRecord
	if err := f.decodeUserField(string(user), fileDBStores, &stores); err != nil {
		return errors.Wrap(err, "remove store")
	}
	for i, s := range stores {
		if s.ID == store {
			essentials.OrderedDelete(&stores, i)
			if err := f.encodeUserField(string(user), fileDBStores, stores); err != nil {
				return errors.Wrap(err, "remove store")
			}
			if err := f.deleteUserField(string(user), f.listField(store)); err != nil {
				return errors.Wrap(err, "remove store")
			}
			return nil
		}
	}
	return errors.New("remove store: store not found")
}

func (f *FileDB) ListEntries(user UserID, store StoreID) ([]*ListEntry, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	var entries []*ListEntry
	if err := f.decodeUserField(string(user), f.listField(store), &entries); err != nil {
		return nil, errors.Wrap(err, "get list entries")
	}

	return entries, nil
}

func (f *FileDB) AddListEntry(user UserID, store StoreID, inf *ListEntryInfo) (ListEntryID, error) {
	uid, err := randomUID()
	if err != nil {
		return "", errors.Wrap(err, "add list entry")
	}
	entryID := ListEntryID(uid)

	f.lock.Lock()
	defer f.lock.Unlock()

	var entries []*ListEntry
	if err := f.decodeUserField(string(user), f.listField(store), &entries); err != nil {
		return "", errors.Wrap(err, "add list entry")
	}
	entries = append(entries, &ListEntry{
		ID:   entryID,
		Info: inf,
	})
	if err := f.encodeUserField(string(user), f.listField(store), &entries); err != nil {
		return "", errors.Wrap(err, "add list entry")
	}
	return entryID, nil
}

func (f *FileDB) RemoveListEntry(user UserID, store StoreID, entry ListEntryID) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	var entries []*ListEntry
	if err := f.decodeUserField(string(user), f.listField(store), &entries); err != nil {
		return errors.Wrap(err, "remove list entry")
	}
	for i, e := range entries {
		if e.ID == entry {
			essentials.OrderedDelete(&entries, i)
			if err := f.encodeUserField(string(user), f.listField(store), &entries); err != nil {
				return errors.Wrap(err, "remove list entry")
			}
			return nil
		}
	}
	return errors.New("remove list entry: entry not found")
}

func (f *FileDB) usernameDir(username string) string {
	nameHash := sha256.Sum256([]byte(username))
	nameStr := base64.StdEncoding.EncodeToString(nameHash[:])
	return filepath.Join(f.Dir, nameStr)
}

func (f *FileDB) listField(storeID StoreID) string {
	nameHash := sha256.Sum256([]byte(storeID))
	nameStr := base64.StdEncoding.EncodeToString(nameHash[:])
	return fileDBListPrefix + nameStr
}

func (f *FileDB) readUserField(username, field string) ([]byte, error) {
	path := filepath.Join(f.usernameDir(username), field)
	return ioutil.ReadFile(path)
}

func (f *FileDB) decodeUserField(username, field string, out interface{}) error {
	data, err := f.readUserField(username, field)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func (f *FileDB) writeUserField(username, field string, data []byte) error {
	path := filepath.Join(f.usernameDir(username), field)
	return ioutil.WriteFile(path, data, 0700)
}

func (f *FileDB) encodeUserField(username, field string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return f.writeUserField(username, field, data)
}

func (f *FileDB) deleteUserField(username, field string) error {
	return os.Remove(filepath.Join(f.usernameDir(username), field))
}

func randomUID() (string, error) {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
