package target

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
)

type StoreDesc interface {
	Name() string
	Address() string
}

type storeDesc struct {
	Info *LocationInfo
}

func (s *storeDesc) Name() string {
	return s.Info.Name()
}

func (s *storeDesc) Address() string {
	return s.Info.Address.AddressLine1
}

type StoreSource struct {
	Client *Client
}

func (s *StoreSource) StoresNear(lat, lon float64) ([]optishop.StoreDesc, error) {
	geocodes, err := s.Client.Geocodes(lat, lon)
	if err != nil {
		return nil, errors.Wrap(err, "stores near")
	}
	if len(geocodes.Locations) == 0 {
		return nil, errors.New("stores near: no locations found")
	}
	return s.QueryStores(geocodes.Locations[0].Address.PostalCode)
}

func (s *StoreSource) QueryStores(query string) ([]optishop.StoreDesc, error) {
	locations, err := s.Client.SearchStores(query)
	if err != nil {
		return nil, err
	}
	var descs []optishop.StoreDesc
	for _, loc := range locations {
		descs = append(descs, &storeDesc{Info: loc})
	}
	return descs, nil
}

func (s *StoreSource) Store(desc optishop.StoreDesc) (optishop.Store, error) {
	return NewStore(strconv.Itoa(desc.(*storeDesc).Info.LocationID))
}

func (s *StoreSource) MarshalStoreDesc(desc optishop.StoreDesc) ([]byte, error) {
	return json.Marshal(desc)
}

func (s *StoreSource) UnmarshalStoreDesc(data []byte) (optishop.StoreDesc, error) {
	var res storeDesc
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "unmarshal store description")
	}
	return &res, nil
}
