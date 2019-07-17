package main

import (
	"github.com/unixpickle/optishop-server/optishop"
)

type ListItem struct {
	Name        string `json:"name"`
	PhotoURL    string `json:"photoUrl"`
	Description string `json:"description"`
	InStock     bool   `json:"inStock"`
	Price       string `json:"price"`
}

func ConvertListItem(prod optishop.InventoryProduct) *ListItem {
	return &ListItem{
		Name:        prod.Name(),
		PhotoURL:    prod.PhotoURL(),
		Description: prod.Description(),
		InStock:     prod.InStock(),
		Price:       prod.Price(),
	}
}

func ConvertListItems(list []optishop.InventoryProduct) []*ListItem {
	res := make([]*ListItem, len(list))
	for i, x := range list {
		res[i] = ConvertListItem(x)
	}
	return res
}

type StoreDesc struct {
	Source  string `json:"source"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Data    []byte `json:"data"`

	Signature string `json:"signature"`
}
