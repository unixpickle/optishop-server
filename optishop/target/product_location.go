package target

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
)

type InStoreLocation struct {
	Block      string `json:"block"`
	Aisle      int    `json:"aisle"`
	Floor      string `json:"floor"`
	Section    int    `json:"section"`
	BlockAisle string `json:"block_aisle"`
}

// ProductDetails contains detailed information about a
// product that doesn't show up in search results.
type ProductDetails struct {
	Product struct {
		Location InStoreLocation `json:"in_store_location"`
		Item     struct {
			ProductClassification struct {
				ProductTypeName string `json:"product_type_name"`
			} `json:"product_classification"`
		} `json:"item"`
	} `json:"product"`
}

// ProductDetails fetches the details of a product.
func (c *Client) ProductDetails(tcin, storeID string) (*ProductDetails, error) {
	q := url.Values{}
	q.Add("excludes", "taxonomy")
	q.Add("key", c.Key())
	q.Add("storeId", storeID)
	u := "https://redsky.target.com/v3/pdp/tcin/" + tcin + "?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "product details")
	}
	var res ProductDetails
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "product details")
	}
	return &res, nil
}

// LocationDetails fetches information about where a
// product is located in a store. Sometimes, this works
// more reliably than Client.ProductDetails().
func LocationDetails(tcin, storeID string) (*InStoreLocation, error) {
	q := url.Values{}
	q.Set("storeId", storeID)
	u := "https://redsky.target.com/v1/location_details/" + tcin + "?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "location details")
	}

	var result struct {
		Product struct {
			InStoreLocation *InStoreLocation `json:"in_store_location"`
		} `json:"product"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrap(err, "location details")
	}
	if result.Product.InStoreLocation == nil {
		return nil, errors.New("location details: missing location in result")
	}
	return result.Product.InStoreLocation, nil
}
