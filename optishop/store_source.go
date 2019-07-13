package optishop

type StoreDesc interface {
	Name() string
	Address() string
}

type StoreSource interface {
	StoresNear(lat, lon float64) ([]StoreDesc, error)
	QueryStores(query string) ([]StoreDesc, error)
	Store(desc StoreDesc) (Store, error)
	MarshalStoreDesc(desc StoreDesc) ([]byte, error)
	UnmarshalStoreDesc(data []byte) (StoreDesc, error)
}
