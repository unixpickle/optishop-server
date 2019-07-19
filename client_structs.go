package main

import "github.com/unixpickle/optishop-server/optishop"

type ClientListItem struct {
	Name        string `json:"name"`
	PhotoURL    string `json:"photoUrl"`
	Description string `json:"description"`
	InStock     bool   `json:"inStock"`
	Price       string `json:"price"`
}

func NewClientListItem(ip optishop.InventoryProduct) *ClientListItem {
	return &ClientListItem{
		Name:        ip.Name(),
		PhotoURL:    ip.PhotoURL(),
		Description: ip.Description(),
		InStock:     ip.InStock(),
		Price:       ip.Price(),
	}
}

type ClientInventoryItem struct {
	*ClientListItem

	Data      []byte `json:"data"`
	Signature string `json:"signature"`
}

type ClientStoreDesc struct {
	ID      string `json:"id,omitempty"`
	Source  string `json:"source"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Data    []byte `json:"data,omitempty"`

	Signature string `json:"signature,omitempty"`
}
