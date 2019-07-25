package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/db"
)

type StoreKeyType int

type StoreIDKeyType int

// StoreKey is the context key used for an optishop.Store.
var StoreKey StoreKeyType

// StoreIDKey is the context key used for a db.StoreID.
var StoreIDKey StoreIDKeyType

// StoreHandler wraps an HTTP handler for requests that
// include a store ID, automatically providing the store
// as part of the context.
func (s *Server) StoreHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(db.UserID)
		storeID := db.StoreID(r.FormValue("store"))

		storeRecord, err := s.DB.Store(user, storeID)
		if err != nil {
			s.ServeError(w, r, err)
			return
		}

		store, err := s.StoreCache.GetStore(storeRecord.Info.SourceName, storeRecord.Info.StoreData)
		if err != nil {
			s.ServeError(w, r, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, StoreKey, store)
		ctx = context.WithValue(ctx, StoreIDKey, storeID)
		h(w, r.WithContext(ctx))
	}
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
