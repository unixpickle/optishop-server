package target

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type FulfillmentResult struct {
	TCIN        string `json:"tcin"`
	Fulfillment struct {
		ShippingOptions struct {
			AvailabilityStatus string `json:"availability_status"`
		} `json:"shipping_options"`
		StoreOptions []struct {
			InStoreOnly struct {
				AvailabilityStatus string `json:"availability_status"`
			} `json:"in_store_only"`
		} `json:"store_options"`
	} `json:"fulfillment"`
	StorePositions []struct {
		Aisle int    `json:"aisle"`
		Block string `json:"block"`
	} `json:"store_positions"`
}

func (f *FulfillmentResult) InStore() bool {
	opts := f.Fulfillment.StoreOptions
	if len(opts) == 0 {
		return false
	}
	return opts[0].InStoreOnly.AvailabilityStatus == "IN_STOCK"
}

func (f *FulfillmentResult) ZoneName() string {
	if len(f.StorePositions) == 0 {
		return ""
	}
	p := f.StorePositions[0]
	return p.Block + strconv.Itoa(p.Aisle)
}

func (c *Client) Fulfillment(storeID string, tcins []string) ([]*FulfillmentResult, error) {
	if len(tcins) == 0 {
		return nil, nil
	}

	q := url.Values{}
	q.Set("store_id", storeID)
	q.Set("scheduled_delivery_store_id", storeID)
	q.Set("channel", "web")
	q.Set("tcins", strings.Join(tcins, ","))
	q.Set("key", c.Key())

	// Dummy values for unused shipping options.
	q.Set("latitude", "0")
	q.Set("longitude", "0")
	q.Set("zip", "19072")
	q.Set("state", "PA")

	u := "https://redsky.target.com/redsky_aggregations/v1/web/plp_fulfillment_v1?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "product fulfillment")
	}

	var response struct {
		Data struct {
			ProductSummaries []*FulfillmentResult `json:"product_summaries"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, errors.Wrap(err, "product fulfillment")
	}
	return response.Data.ProductSummaries, nil
}

// SingleFulfillment gets fulfillment info for a single
// product. Unlike Fulfillment(), the product's store
// position should be available in the result.
func (c *Client) SingleFulfillment(storeID string, tcin string) (*FulfillmentResult, error) {
	q := url.Values{}
	q.Set("store_id", storeID)
	q.Set("scheduled_delivery_store_id", storeID)
	q.Set("store_positions_store_id", storeID)
	q.Set("pricing_store_id", storeID)
	q.Set("has_store_positions_store_id", "true")
	q.Set("has_pricing_store_id", "true")
	q.Set("channel", "web")
	q.Set("tcin", tcin)
	q.Set("key", c.Key())

	// Dummy values for unused shipping options.
	q.Set("latitude", "0")
	q.Set("longitude", "0")
	q.Set("zip", "19072")
	q.Set("state", "PA")

	u := "https://redsky.target.com/redsky_aggregations/v1/web/pdp_fulfillment_v1?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "single fulfillment")
	}

	var response struct {
		Data struct {
			Product *FulfillmentResult `json:"product"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, errors.Wrap(err, "single fulfillment")
	}
	if response.Data.Product == nil {
		return nil, errors.New("single fulfillment: missing product in result")
	}
	return response.Data.Product, nil
}
