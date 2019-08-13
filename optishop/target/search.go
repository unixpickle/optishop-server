package target

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PackageDimensions struct {
	Weight        string `json:"weight"`
	WeightUnit    string `json:"weight_unit_of_measure"`
	Width         string `json:"width"`
	Height        string `json:"height"`
	Depth         string `json:"depth"`
	DimensionUnit string `json:"dimension_unit_of_measure"`
}

type ProductImage struct {
	BaseURL       string   `json:"base_url"`
	Primary       string   `json:"primary"`
	AlternateURLs []string `json:"alternate_urls"`
}

type ProductPrice struct {
	TCIN                      string  `json:"tcin"`
	FormattedCurrentPrice     string  `json:"formatted_current_price"`
	FormattedCurrentPriceType string  `json:"formatted_current_price_type"`
	IsCurrentPriceRange       bool    `json:"is_current_price_range"`
	CurrentRetail             float64 `json:"current_retail"`
}

// A SearchItem is a single result from a product search.
type SearchItem struct {
	Subscribable                  bool               `json:"subscribable"`
	PackageDimensions             *PackageDimensions `json:"package_dimensions"`
	Title                         string             `json:"title"`
	TCIN                          string             `json:"tcin"`
	Type                          string             `json:"type"`
	DPCI                          string             `json:"dpci"`
	UPC                           string             `json:"upc"`
	URL                           string             `json:"url"`
	Description                   string             `json:"description"`
	RepresentativeChildPartNumber string             `json:"representative_child_part_number"`
	MerchSubClass                 string             `json:"merch_sub_class"`
	MerchClass                    string             `json:"merch_class"`
	MerchClassID                  string             `json:"merch_class_id"`
	Brand                         string             `json:"brand"`
	ProductBrand                  struct {
		FacetID string `json:"facet_id"`
		Brand   string `json:"brand"`
	} `json:"product_brand"`
	Images                []*ProductImage `json:"images"`
	AvailabilityStatus    string          `json:"availability_status"`
	SDSAvailabilityStatus string          `json:"scheduled_delivery_store_availability_status"`
	Price                 ProductPrice    `json:"price"`
	BulletDescription     []string        `json:"bullet_description"`
}

// SearchResults stores results from a product search.
type SearchResults struct {
	Items struct {
		SearchItems []*SearchItem `json:"Item"`
	} `json:"items"`
	Metadata    *RequestMetadata `json:"metaData"`
	Suggestions []string         `json:"suggestions"`
}

// Search runs a search query against a store's inventory.
func (c *Client) Search(query, storeID string, offset int) (*SearchResults, error) {
	q := url.Values{}
	q.Add("channel", "web")
	q.Add("count", "24")
	q.Add("default_purchasability_filter", "true")
	q.Add("isDLP", "false")
	q.Add("keyword", strings.ToLower(query))
	q.Add("offset", strconv.Itoa(offset))
	q.Add("pricing_store_id", storeID)
	q.Add("scheduled_delivery_store_id", storeID)
	q.Add("store_ids", storeID)
	q.Add("include_sponsored_search", "false")
	q.Add("platform", "desktop")
	q.Add("key", c.Key())
	u := "https://redsky.target.com/v2/plp/search/?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "search")
	}
	var res struct {
		Results SearchResults `json:"search_response"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "search")
	}
	return &res.Results, nil
}
