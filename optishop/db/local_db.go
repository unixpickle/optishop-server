package db

import "github.com/pkg/errors"

// A LocalDB is like a FileDB, but there is only one
// logical user and all requests are automatically pushed
// through this user.
type LocalDB struct {
	fileDB *FileDB
	userID UserID
}

// NewLocalDB creates a LocalDB at the given directory
// path, creating the directory if necessary.
func NewLocalDB(path string) (*LocalDB, error) {
	fdb, err := NewFileDB(path)
	if err != nil {
		return nil, errors.Wrap(err, "create local DB")
	}
	userID, err := fdb.Login("local", "local")
	if err != nil {
		userID, err = fdb.CreateUser("local", "local", map[string]string{})
	}
	if err != nil {
		return nil, errors.Wrap(err, "create local DB")
	}
	return &LocalDB{
		fileDB: fdb,
		userID: userID,
	}, nil
}

func (l *LocalDB) CreateUser(username, password string, metadata map[string]string) (UserID, error) {
	return "", errors.New("create user: not implemented")
}

func (l *LocalDB) Chpass(user UserID, old, new string) error {
	return errors.New("change password: not implemented")
}

func (l *LocalDB) Login(username, password string) (UserID, error) {
	return "", nil
}

func (l *LocalDB) Username(user UserID) (string, error) {
	return "local", nil
}

func (l *LocalDB) UserMetadata(user UserID, field string) (string, error) {
	return l.fileDB.UserMetadata(l.userID, field)
}

func (l *LocalDB) SetUserMetadata(user UserID, field, value string) error {
	return l.fileDB.SetUserMetadata(l.userID, field, value)
}

func (l *LocalDB) Stores(user UserID) ([]*StoreRecord, error) {
	return l.fileDB.Stores(l.userID)
}

func (l *LocalDB) Store(user UserID, store StoreID) (*StoreRecord, error) {
	return l.fileDB.Store(l.userID, store)
}

func (l *LocalDB) AddStore(user UserID, info *StoreInfo) (StoreID, error) {
	return l.fileDB.AddStore(l.userID, info)
}

func (l *LocalDB) RemoveStore(user UserID, store StoreID) error {
	return l.fileDB.RemoveStore(l.userID, store)
}

func (l *LocalDB) ListEntries(user UserID, store StoreID) ([]*ListEntry, error) {
	return l.fileDB.ListEntries(l.userID, store)
}

func (l *LocalDB) AddListEntry(user UserID, store StoreID, info *ListEntryInfo) (ListEntryID, error) {
	return l.fileDB.AddListEntry(l.userID, store, info)
}

func (l *LocalDB) RemoveListEntry(user UserID, store StoreID, entry ListEntryID) error {
	return l.fileDB.RemoveListEntry(l.userID, store, entry)
}

func (l *LocalDB) PermuteListEntries(user UserID, store StoreID, ids []ListEntryID) error {
	return l.fileDB.PermuteListEntries(l.userID, store, ids)
}
