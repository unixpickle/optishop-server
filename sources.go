package main

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/target"
)

// LoadStoreSources creates a map of supported store
// sources.
func LoadStoreSources() (map[string]optishop.StoreSource, error) {
	client, err := target.NewClient()
	if err != nil {
		return nil, err
	}
	return map[string]optishop.StoreSource{
		"target": &target.StoreSource{Client: client},
	}, nil
}

// A StoreCache uses a cache to quickly retrieve Store
// objects for serialized store descriptions.
type StoreCache struct {
	sources map[string]optishop.StoreSource

	lock  sync.RWMutex
	cache map[cacheKey]optishop.Store
}

// NewStoreCache creates a StoreCache that will used the
// named collection of store sources.
func NewStoreCache(sources map[string]optishop.StoreSource) *StoreCache {
	return &StoreCache{
		sources: sources,
		cache:   map[cacheKey]optishop.Store{},
	}
}

// GetStore looks up a Store for the source name and the
// serialized optishop.StoreDesc.
func (s *StoreCache) GetStore(sourceName string, descData []byte) (optishop.Store, error) {
	source, ok := s.sources[sourceName]
	if !ok {
		return nil, errors.New("get store: no such source: " + sourceName)
	}
	desc, err := source.UnmarshalStoreDesc(descData)
	if err != nil {
		return nil, errors.Wrap(err, "get store")
	}

	key := cacheKey{Source: sourceName, Name: desc.Name(), Address: desc.Address()}
	s.lock.RLock()
	existing, ok := s.cache[key]
	s.lock.RUnlock()
	if ok {
		return existing, nil
	}

	store, err := source.Store(desc)
	if err != nil {
		return nil, errors.Wrap(err, "get store")
	}

	s.lock.Lock()
	s.cache[key] = store
	s.lock.Unlock()

	return store, nil
}

type cacheKey struct {
	Source  string
	Name    string
	Address string
}
