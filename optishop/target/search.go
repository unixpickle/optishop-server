package target

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
	Subscribable      bool               `json:"subscribable"`
	PackageDimensions *PackageDimensions `json:"package_dimensions"`
	Title             string             `json:"title"`
	TCIN              string             `json:"tcin"`
	Type              string             `json:"type"`
	DPCI              string             `json:"dpci"`
	UPC               string             `json:"upc"`
	URL               string             `json:"url"`
	Description       string             `json:"description"`
	MerchSubClass     string             `json:"merch_sub_class"`
	MerchClass        string             `json:"merch_class"`
	MerchClassID      string             `json:"merch_class_id"`
	Brand             string             `json:"brand"`
	ProductBrand      struct {
		FacetID string `json:"facet_id"`
		Brand   string `json:"brand"`
	} `json:"product_brand"`
	Images             []*ProductImage `json:"images"`
	AvailabilityStatus string          `json:"availability_status"`
	Price              *ProductPrice   `json:"price"`
}

// SearchResults stores results from a product search.
type SearchResults struct {
	Items struct {
		SearchItems []*SearchItem `json:"Item"`
	} `json:"items"`
	Metadata    *RequestMetadata `json:"metaData"`
	Suggestions []string         `json:"suggestions"`
}
