package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"unicode"

	"github.com/unixpickle/essentials"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	fileDBHash       = "hash"
	fileDBStores     = "stores"
	fileDBUsername   = "username"
	fileDBListPrefix = "store_"
	fileDBMeta       = "meta_"
)

// A FileDB uses the filesystem for an extremely simple
// database.
type FileDB struct {
	Dir  string
	lock sync.RWMutex
}

// NewFileDB creates a FileDB at the given directory path,
// creating the directory if necessary.
func NewFileDB(path string) (*FileDB, error) {
	if err := os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &FileDB{Dir: path}, nil
}

func (f *FileDB) CreateUser(username, password string, metadata map[string]string) (UserID, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	userDir := f.usernameDir(username)
	if err := os.Mkdir(userDir, 0700); err != nil {
		if os.IsExist(err) {
			return "", errors.New("create user: user already exists")
		}
		return "", errors.Wrap(err, "create user")
	}

	if err := f.setupNewUserFields(username, password, metadata); err != nil {
		os.RemoveAll(userDir)
		return "", errors.Wrap(err, "create user")
	}

	return UserID(username), nil
}

func (f *FileDB) setupNewUserFields(username, password string, metadata map[string]string) error {
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
	if err := f.writeUserField(username, fileDBUsername, []byte(username)); err != nil {
		return err
	}
	for field, value := range metadata {
		if err := validateMetadataFieldName(field); err != nil {
			return err
		}
		if err := f.writeUserField(username, fileDBMeta+field, []byte(value)); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileDB) Chpass(user UserID, old, new string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	hash, err := f.readUserField(string(user), fileDBHash)
	if err != nil {
		return errors.Wrap(err, "change password")
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(old)) != nil {
		return errors.New("change password: incorrect old password")
	}

	hash, err = bcrypt.GenerateFromPassword([]byte(new), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "change password")
	}
	if err := f.writeUserField(string(user), fileDBHash, hash); err != nil {
		return errors.Wrap(err, "change password")
	}

	return nil
}

func (f *FileDB) UserMetadata(user UserID, field string) (string, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if err := validateMetadataFieldName(field); err != nil {
		return "", errors.Wrap(err, "get user metadata field")
	}
	data, err := f.readUserField(string(user), fileDBMeta+field)
	if err != nil {
		return "", errors.Wrap(err, "get user metadata field")
	}
	return string(data), nil
}

func (f *FileDB) SetUserMetadata(user UserID, field, value string) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if err := validateMetadataFieldName(field); err != nil {
		return errors.Wrap(err, "set user metadata field")
	}
	if err := f.writeUserField(string(user), fileDBMeta+field, []byte(value)); err != nil {
		return errors.Wrap(err, "set user metadata field")
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

func (f *FileDB) Store(user UserID, store StoreID) (*StoreRecord, error) {
	stores, err := f.Stores(user)
	if err != nil {
		return nil, errors.Wrap(err, "get store")
	}
	for _, x := range stores {
		if x.ID == store {
			return x, nil
		}
	}
	return nil, errors.New("get store: store not found")
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

func (f *FileDB) PermuteListEntries(user UserID, store StoreID, ids []ListEntryID) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	var entries []*ListEntry
	if err := f.decodeUserField(string(user), f.listField(store), &entries); err != nil {
		return errors.Wrap(err, "permute list entries")
	}

	if len(ids) != len(entries) {
		return errors.New("permute list entries: entries have changed")
	}

	mapping := map[ListEntryID]*ListEntry{}
	for _, entry := range entries {
		mapping[entry.ID] = entry
	}

	newEntries := make([]*ListEntry, len(ids))
	for i, id := range ids {
		if entry, ok := mapping[id]; !ok {
			return errors.New("permute list entries: entry not found or duplicate ID")
		} else {
			newEntries[i] = entry
			delete(mapping, id)
		}
	}

	if err := f.encodeUserField(string(user), f.listField(store), newEntries); err != nil {
		return errors.Wrap(err, "permute list entries")
	}

	return nil
}

func (f *FileDB) usernameDir(username string) string {
	nameHash := sha256.Sum256([]byte(username))
	nameStr := base64.URLEncoding.EncodeToString(nameHash[:])
	return filepath.Join(f.Dir, nameStr)
}

func (f *FileDB) listField(storeID StoreID) string {
	nameHash := sha256.Sum256([]byte(storeID))
	nameStr := base64.URLEncoding.EncodeToString(nameHash[:])
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
	return base64.URLEncoding.EncodeToString(data), nil
}

func validateMetadataFieldName(name string) error {
	for _, x := range []rune(name) {
		if !unicode.IsLetter(x) && !unicode.IsNumber(x) && x != '_' {
			return fmt.Errorf("disallowed character in metadata field: %c", x)
		}
	}
	return nil
}
