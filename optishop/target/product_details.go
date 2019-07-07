package target

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
)

// ProductDetails contains detailed information about a
// product that doesn't show up in search results.
type ProductDetails struct {
	Product struct {
		Location struct {
			Block      string `json:"block"`
			Aisle      int    `json:"aisle"`
			Floor      string `json:"floor"`
			Section    int    `json:"section"`
			BlockAisle string `json:"block_aisle"`
		} `json:"in_store_location"`
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
