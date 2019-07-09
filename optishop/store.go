package optishop

type InventoryProduct interface {
	Name() string
	PhotoURL() string
	Description() string
	InStock() bool
	Price() string
}

type Inventory interface {
	Search(query string) ([]InventoryProduct, error)

	MarshalProduct(prod InventoryProduct) ([]byte, error)
	UnmarshalProduct(data []byte) (InventoryProduct, error)
}

type Store interface {
	Inventory

	// Layout gets the layout of the store.
	Layout() *Layout

	// Locate gets the zone containing a product.
	//
	// Returns nil if the product's location is unknown.
	Locate(product InventoryProduct) (*Zone, error)
}
