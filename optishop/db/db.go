// Package db implements the database for the backend of
// the optishop server.package db
package db

import "github.com/unixpickle/optishop-server/optishop"

type UserID interface{}

type StoreID interface{}

type ListEntryID interface{}

type StoreRecord struct {
	ID   StoreID
	Info *StoreInfo
}

type StoreInfo struct {
	SourceName string
	StoreName  string
	StoreData  []byte
}

type ListEntry struct {
	ID   ListEntryID
	Info *ListEntryInfo
}

type ListEntryInfo struct {
	InventoryProductData []byte
	Zone                 *optishop.Zone
}

type DB interface {
	Login(username, password string) (UserID, error)
	Stores(user UserID) ([]*StoreInfo, error)
	AddStore(user UserID, info *StoreInfo) error
	RemoveStore(user UserID, store StoreID) error
	ListEntries(user UserID, store StoreID) ([]*ListEntry, error)
	AddListEntry(user UserID, store StoreID, info *ListEntryInfo) error
	RemoveListEntry(user UserID, store StoreID, entry ListEntryID) error
}
