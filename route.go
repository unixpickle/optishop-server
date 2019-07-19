package main

import (
	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/db"
)

// SortEntries finds the optimal route and returns the
// list entries sorted by this route.
func SortEntries(list []*db.ListEntry, store optishop.Store) ([]*db.ListEntry, error) {
	// TODO: this.
	return nil, errors.New("sort entries: not yet implemented")
}
