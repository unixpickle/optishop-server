// Package db implements the database for the backend of
// the optishop server.
package db

import "github.com/unixpickle/optishop-server/optishop"

type UserID string

type StoreID string

type ListEntryID string

type StoreRecord struct {
	ID   StoreID
	Info *StoreInfo
}

type StoreInfo struct {
	SourceName   string
	StoreName    string
	StoreAddress string
	StoreData    []byte
}

type ListEntry struct {
	ID   ListEntryID
	Info *ListEntryInfo
}

type ListEntryInfo struct {
	InventoryProductData []byte
	Zone                 *optishop.Zone
	Floor                int
}

type DB interface {
	CreateUser(username, password string, metadata map[string]string) (UserID, error)
	Chpass(user UserID, old, new string) error
	Login(username, password string) (UserID, error)
	UserMetadata(user UserID, field string) (string, error)
	SetUserMetadata(user UserID, field, value string) error
	Stores(user UserID) ([]*StoreRecord, error)
	Store(user UserID, store StoreID) (*StoreRecord, error)
	AddStore(user UserID, info *StoreInfo) (StoreID, error)
	RemoveStore(user UserID, store StoreID) error
	ListEntries(user UserID, store StoreID) ([]*ListEntry, error)
	AddListEntry(user UserID, store StoreID, info *ListEntryInfo) (ListEntryID, error)
	RemoveListEntry(user UserID, store StoreID, entry ListEntryID) error
	PermuteListEntries(user UserID, store StoreID, ids []ListEntryID) error
}
