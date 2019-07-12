package target

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

type GeocodesResult struct {
	Place     string `json:"place"`
	Locale    string `json:"locale"`
	Locations []struct {
		Type    string `json:"type"`
		Address struct {
			Latitude         float64 `json:"latitude"`
			Longitude        float64 `json:"longitude"`
			City             string  `json:"city"`
			Subdivision      string  `json:"subdivision"`
			PostalCode       string  `json:"postalCode"`
			CountryName      string  `json:"countryName"`
			FormattedAddress string  `json:"formattedAddress"`
		} `json:"address"`
	} `json:"locations"`
}

// SearchStores finds stores based on a query, e.g. a zip
// code, address, state, etc.
func (c *Client) SearchStores(query string) {
	// https://redsky.target.com/v3/stores/nearby/georgia?key=eb2551e4accc14f38cc42d32fbc2b2ea&limit=20&within=100&unit=mile
}

// Geocodes searches a lat/lon pair and yields address
// data which can then be fed into SearchStores.
func (c *Client) Geocodes(lat, lon float64) (*GeocodesResult, error) {
	q := url.Values{}
	q.Add("key", c.Key())
	q.Add("place", fmt.Sprintf("%f,%f", lat, lon))
	u := "https://api.target.com/location_proximities/v1/geocodes?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "get geocodes")
	}
	var res GeocodesResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, errors.Wrap(err, "get geocodes")
	}
	return &res, nil
}
