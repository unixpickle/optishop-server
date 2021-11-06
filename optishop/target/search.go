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
//
// This is a legacy data structure and is no longer
// returned by search results.
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
//
// This is a legacy data structure and is no longer
// returned by search results.
type SearchResults struct {
	Items struct {
		SearchItems []*SearchItem `json:"Item"`
	} `json:"items"`
	Metadata    *RequestMetadata `json:"metaData"`
	Suggestions []string         `json:"suggestions"`
}

type SearchProduct struct {
	TCIN  string       `json:"tcin"`
	Price ProductPrice `json:"price"`
	Item  struct {
		Enrichment struct {
			Images struct {
				PrimaryURL string `json:"primary_image_url"`
			} `json:"images"`
		} `json:"enrichment"`
		DPCI        string `json:"dpci"`
		Description struct {
			Title              string   `json:"title"`
			BulletDescriptions []string `json:"bullet_descriptions"`
			SoftBullets        struct {
				Bullets []string `json:"bullets"`
			} `json:"soft_bullets"`
		} `json:"product_description"`
	} `json:"item"`
}

type SearchResultsV2 struct {
	Data struct {
		Search struct {
			Response struct {
				Metadata struct {
					Count        int `json:"count"`
					Offset       int `json:"offset"`
					TotalResults int `json:"total_results"`
				} `json:"typed_metadata"`
			} `json:"search_response"`
			Suggestions []string         `json:"search_suggestions"`
			Products    []*SearchProduct `json:"products"`
		} `json:"search"`
	} `json:"data"`
}

// Search runs a search query against a store's inventory.
func (c *Client) Search(query, storeID string, offset int) (*SearchResultsV2, error) {
	q := url.Values{}
	q.Add("channel", "WEB")
	q.Add("count", "24")
	q.Add("default_purchasability_filter", "true")
	q.Add("include_sponsored", "false")
	q.Add("keyword", strings.ToLower(query))
	q.Add("offset", strconv.Itoa(offset))
	q.Add("page", "/s/"+url.PathEscape(query))
	q.Add("pricing_store_id", storeID)
	q.Add("scheduled_delivery_store_id", storeID)
	q.Add("store_ids", storeID)
	q.Add("platform", "desktop")
	q.Add("key", c.Key())
	q.Add("visitor_id", c.VisitorID())
	u := "https://redsky.target.com/redsky_aggregations/v1/web/plp_search_v1?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "search")
	}
	var res SearchResultsV2
	if len(strings.TrimSpace(string(data))) == 0 {
		return &res, nil
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "search")
	}
	return &res, nil
}
