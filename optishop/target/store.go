package target

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"golang.org/x/net/html"
)

type inventoryProduct struct {
	SearchItem  *SearchItem
	ShipMethods *ShipMethodsResult
}

func (i *inventoryProduct) Name() string {
	return html.UnescapeString(i.SearchItem.Title)
}

func (i *inventoryProduct) PhotoURL() string {
	for _, img := range i.SearchItem.Images {
		return img.BaseURL + img.Primary
	}
	return ""
}

func (i *inventoryProduct) Description() string {
	res := html.UnescapeString(i.SearchItem.Description)
	if res != "" {
		return cleanupDescriptionHTML(res)
	}
	lines := []string{}
	for _, line := range i.SearchItem.BulletDescription {
		lines = append(lines, "- "+cleanupDescriptionHTML(line))
	}
	return strings.Join(lines, "\n")
}

func cleanupDescriptionHTML(desc string) string {
	replacer := regexp.MustCompilePOSIX("<[^>]*>")
	cleaned := desc
	cleaned = strings.ReplaceAll(cleaned, "<br>", "\n")
	cleaned = strings.ReplaceAll(cleaned, "<br />", "\n")
	cleaned = replacer.ReplaceAllString(cleaned, "")
	cleaned = html.UnescapeString(cleaned)
	return cleaned
}

func (i *inventoryProduct) InStock() bool {
	return i.ShipMethods != nil && i.ShipMethods.InStore()
}

func (i *inventoryProduct) Price() string {
	p := i.SearchItem.Price
	if strings.HasPrefix(p.FormattedCurrentPrice, "$") {
		return p.FormattedCurrentPrice
	} else {
		return fmt.Sprintf("$%0.2f", p.CurrentRetail)
	}
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

func (s *Store) Search(query string) ([]optishop.InventoryProduct, []string, error) {
	results, err := s.Client.Search(query, s.StoreID, 0)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]string, len(results.Items.SearchItems))
	for i, result := range results.Items.SearchItems {
		ids[i] = result.RepresentativeChildPartNumber
	}
	shipMethods, err := ShipMethods(s.StoreID, ids)
	if err != nil {
		return nil, nil, err
	}
	idToShipMethod := map[string]*ShipMethodsResult{}
	for _, method := range shipMethods {
		idToShipMethod[method.ProductID] = method
	}

	var products []optishop.InventoryProduct
	for _, res := range results.Items.SearchItems {
		products = append(products, &inventoryProduct{
			SearchItem:  res,
			ShipMethods: idToShipMethod[res.RepresentativeChildPartNumber],
		})
	}

	return products, results.Suggestions, nil
}

func (s *Store) MarshalProduct(prod optishop.InventoryProduct) ([]byte, error) {
	return json.Marshal(prod)
}

func (s *Store) UnmarshalProduct(data []byte) (optishop.InventoryProduct, error) {
	var prod inventoryProduct
	if err := json.Unmarshal(data, &prod); err != nil {
		return nil, errors.Wrap(err, "unmarshal product")
	}
	return &prod, nil
}

func (s *Store) Layout() *optishop.Layout {
	return s.CachedLayout
}

func (s *Store) Locate(prod optishop.InventoryProduct) (*optishop.Zone, error) {
	item := prod.(*inventoryProduct).SearchItem

	if loc, err := LocationDetails(item.RepresentativeChildPartNumber, s.StoreID); err == nil {
		name := strings.Replace(loc.BlockAisle, "-", "", -1)
		zone := s.CachedLayout.Zone(name)
		if zone == nil {
			return nil, errors.New("locate product: aisle " + name + " is missing from the map")
		}
		return zone, nil
	}

	details, err := s.Client.ProductDetails(item.RepresentativeChildPartNumber, s.StoreID)
	if err != nil {
		return nil, err
	}
	name := strings.ToLower(details.Product.Item.ProductClassification.ProductTypeName)
	return s.CachedLayout.Zone(name), nil
}
