package main

import (
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/target"
)

func LoadStoreSources() (map[string]optishop.StoreSource, error) {
	client, err := target.NewClient()
	if err != nil {
		return nil, err
	}
	return map[string]optishop.StoreSource{
		"target": &target.StoreSource{Client: client},
	}, nil
}
