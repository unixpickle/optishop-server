package main

type ListItem struct {
	Name        string `json:"name"`
	PhotoURL    string `json:"photoUrl"`
	Description string `json:"description"`
	InStock     bool   `json:"inStock"`
	Price       string `json:"price"`
}

type StoreDesc struct {
	Source  string `json:"source"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Data    []byte `json:"data"`

	Signature string `json:"signature"`
}
