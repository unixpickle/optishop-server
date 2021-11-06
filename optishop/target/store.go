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

type inventoryProductV1 struct {
	SearchItem  *SearchItem
	ShipMethods *ShipMethodsResult
}

func (i *inventoryProductV1) Name() string {
	return html.UnescapeString(i.SearchItem.Title)
}

func (i *inventoryProductV1) PhotoURL() string {
	for _, img := range i.SearchItem.Images {
		return img.BaseURL + img.Primary
	}
	return ""
}

func (i *inventoryProductV1) Description() string {
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

func (i *inventoryProductV1) InStock() bool {
	return i.ShipMethods != nil && i.ShipMethods.InStore()
}

func (i *inventoryProductV1) Price() string {
	p := i.SearchItem.Price
	if strings.HasPrefix(p.FormattedCurrentPrice, "$") {
		return p.FormattedCurrentPrice
	} else {
		return fmt.Sprintf("$%0.2f", p.CurrentRetail)
	}
}

func (i *inventoryProductV1) TCIN() string {
	return i.SearchItem.TCIN
}

type inventoryProductV2 struct {
	SearchProduct *SearchProduct
	Fulfillment   *FulfillmentResult
}

func (i *inventoryProductV2) Name() string {
	return html.UnescapeString(i.SearchProduct.Item.Description.Title)
}

func (i *inventoryProductV2) PhotoURL() string {
	return i.SearchProduct.Item.Enrichment.Images.PrimaryURL
}

func (i *inventoryProductV2) Description() string {
	lines := append([]string{}, i.SearchProduct.Item.Description.BulletDescriptions...)
	lines = append(lines, i.SearchProduct.Item.Description.SoftBullets.Bullets...)
	dashed := []string{}
	for _, line := range lines {
		dashed = append(dashed, "- "+cleanupDescriptionHTML(line))
	}
	return strings.Join(dashed, "\n")
}

func (i *inventoryProductV2) InStock() bool {
	return i.Fulfillment != nil && i.Fulfillment.InStore()
}

func (i *inventoryProductV2) Price() string {
	p := i.SearchProduct.Price
	if strings.HasPrefix(p.FormattedCurrentPrice, "$") {
		return p.FormattedCurrentPrice
	} else {
		return fmt.Sprintf("$%0.2f", p.CurrentRetail)
	}
}

func (i *inventoryProductV2) TCIN() string {
	return i.SearchProduct.TCIN
}

type tcinItem interface {
	TCIN() string
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

	ids := make([]string, len(results.Data.Search.Products))
	for i, result := range results.Data.Search.Products {
		ids[i] = result.TCIN
	}
	fulfillmentResults, err := s.Client.Fulfillment(s.StoreID, ids)
	if err != nil {
		return nil, nil, err
	}
	idToFulfillment := map[string]*FulfillmentResult{}
	for _, res := range fulfillmentResults {
		idToFulfillment[res.TCIN] = res
	}

	var products []optishop.InventoryProduct
	for _, res := range results.Data.Search.Products {
		products = append(products, &inventoryProductV2{
			SearchProduct: res,
			Fulfillment:   idToFulfillment[res.TCIN],
		})
	}

	return products, results.Data.Search.Suggestions, nil
}

func (s *Store) MarshalProduct(prod optishop.InventoryProduct) ([]byte, error) {
	return json.Marshal(prod)
}

func (s *Store) UnmarshalProduct(data []byte) (optishop.InventoryProduct, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, errors.Wrap(err, "unmarshal product")
	}
	if _, ok := obj["SearchItem"]; ok {
		var prod inventoryProductV1
		if err := json.Unmarshal(data, &prod); err != nil {
			return nil, errors.Wrap(err, "unmarshal product")
		}
		return &prod, nil
	} else {
		var prod inventoryProductV2
		if err := json.Unmarshal(data, &prod); err != nil {
			return nil, errors.Wrap(err, "unmarshal product")
		}
		return &prod, nil
	}
}

func (s *Store) Layout() *optishop.Layout {
	return s.CachedLayout
}

func (s *Store) Locate(prod optishop.InventoryProduct) (*optishop.Zone, error) {
	tcin := prod.(tcinItem).TCIN()
	if res, err := s.Client.SingleFulfillment(s.StoreID, tcin); err != nil {
		return nil, errors.Wrap(err, "locate product")
	} else {
		name := res.ZoneName()
		zone := s.CachedLayout.Zone(name)
		if zone == nil {
			return nil, errors.New("locate product: position " + name + " is missing from the map")
		}
		return zone, nil
	}
}
