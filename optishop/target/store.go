package target

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
)

type inventoryProduct struct {
	SearchItem *SearchItem
}

func (i *inventoryProduct) Name() string {
	return i.SearchItem.Title
}

func (i *inventoryProduct) PhotoURL() string {
	for _, img := range i.SearchItem.Images {
		return img.BaseURL + img.Primary
	}
	return ""
}

func (i *inventoryProduct) Description() string {
	return i.SearchItem.Description
}

func (i *inventoryProduct) InStock() bool {
	return i.SearchItem.AvailabilityStatus == "IN_STOCK"
}

type Store struct {
	StoreID      string
	Client       *Client
	CachedLayout *optishop.Layout
}

func NewStore(storeID string) (*Store, error) {
	mapInfo, err := GetMapInfo(storeID)
	if err != nil {
		return nil, errors.Wrap(err, "create store")
	}

	layout := MapInfoToLayout(mapInfo)

	for i, floor := range layout.Floors {
		floorDetails, err := GetFloorDetails(storeID, mapInfo.FloorMapDetails[i].FloorID)
		if err != nil {
			return nil, errors.Wrap(err, "create store")
		}
		AddFloorDetails(floorDetails, floor)
	}

	client, err := NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "create store")
	}

	return &Store{
		StoreID:      storeID,
		Client:       client,
		CachedLayout: layout,
	}, nil
}

func (s *Store) Search(query string) ([]optishop.InventoryProduct, error) {
	results, err := s.Client.Search(query, s.StoreID, 0)
	if err != nil {
		return nil, err
	}
	var products []optishop.InventoryProduct
	for _, res := range results.Items.SearchItems {
		products = append(products, &inventoryProduct{SearchItem: res})
	}
	return products, nil
}

func (s *Store) Layout() *optishop.Layout {
	return s.CachedLayout
}

func (s *Store) Locate(prod optishop.InventoryProduct) (*optishop.Zone, error) {
	item := prod.(*inventoryProduct).SearchItem
	details, err := s.Client.ProductDetails(item.TCIN, s.StoreID)
	if err != nil {
		return nil, err
	}
	zoneName := strings.Replace(details.Product.Location.BlockAisle, "-", "", -1)
	return s.CachedLayout.Zone(zoneName), nil
}