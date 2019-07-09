package target

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// ShipMethodsResult is the result for a single product in
// a result from calling ShipMethods.
type ShipMethodsResult struct {
	ProductID             string `json:"product_id"`
	AvailabilityStatus    string `json:"availability_status"`
	PreferredStoreOptions []struct {
		PreferredStoreOptionID string `json:"preferred_store_option_id"`
		LocationID             string `json:"location_id"`
	} `json:"preferred_store_options"`
}

// InStore checks if the corresponding item is available
// in stores.
func (s *ShipMethodsResult) InStore() bool {
	for _, opt := range s.PreferredStoreOptions {
		if opt.PreferredStoreOptionID == "IN_STORE" {
			return true
		}
	}
	return false
}

// ShipMethods gets shipping information for a list of
// products.
//
// Each product is identified by its representative child
// part number.
func ShipMethods(storeID string, partIDs []string) ([]*ShipMethodsResult, error) {
	joined := strings.Join(partIDs, ",")
	q := url.Values{}

	// Dummy values for unused shipping options.
	q.Set("latitude", "0")
	q.Set("longitude", "0")
	q.Set("zip", "19072")
	q.Set("state", "PA")

	q.Set("storeId", storeID)
	q.Set("channel", "web")

	u := "https://redsky.target.com/v1/ship_methods/aggregator/" + joined + "?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "ship methods")
	}

	var response []struct {
		Product struct {
			ShipMethods *ShipMethodsResult `json:"ship_methods"`
		} `json:"product"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, errors.Wrap(err, "ship methods")
	}

	var results []*ShipMethodsResult
	for _, obj := range response {
		results = append(results, obj.Product.ShipMethods)
	}

	return results, nil
}
