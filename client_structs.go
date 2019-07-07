package main

import (
	"github.com/unixpickle/optishop-server/optishop"
)

type ListItem struct {
	Name        string `json:"name"`
	PhotoURL    string `json:"photo_url"`
	Description string `json:"description"`
	InStock     bool   `json:"in_stock"`
}

func ConvertListItem(prod optishop.InventoryProduct) *ListItem {
	return &ListItem{
		Name:        prod.Name(),
		PhotoURL:    prod.PhotoURL(),
		Description: prod.Description(),
		InStock:     prod.InStock(),
	}
}

func ConvertListItems(list []optishop.InventoryProduct) []*ListItem {
	res := make([]*ListItem, len(list))
	for i, x := range list {
		res[i] = ConvertListItem(x)
	}
	return res
}
