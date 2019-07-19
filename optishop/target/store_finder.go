package target

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// A GeocodesResult stores the result of a Geocodes call.
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

// LocationInfo stores information about a store location
// from a SearchStores query.
type LocationInfo struct {
	LocationID      int    `json:"location_id"`
	TypeCode        string `json:"type_code"`
	TypeDescription string `json:"type_description"`
	LocationNames   []struct {
		NameType string `json:"name_type"`
		Name     string `json:"name"`
	} `json:"location_names"`
	Address struct {
		AddressLine1 string `json:"address_line1"`
		City         string `json:"city"`
		County       string `json:"county"`
		Region       string `json:"region"`
		State        string `json:"state"`
		PostalCode   string `json:"postal_code"`
	}
	Distance     float64 `json:"distance"`
	DistanceUnit string  `json:"distance_unit"`
}

// Name gets the human-readable name of the store.
func (l *LocationInfo) Name() string {
	for _, name := range l.LocationNames {
		if name.NameType == "Proj Name" {
			return name.Name
		}
	}
	if len(l.LocationNames) > 0 {
		return l.LocationNames[0].Name
	}
	return fmt.Sprintf("Location #%d", l.LocationID)
}

// SearchStores finds stores based on a query, e.g. a zip
// code, address, state, etc.
func (c *Client) SearchStores(query string) ([]*LocationInfo, error) {
	q := url.Values{}
	q.Add("key", c.Key())
	q.Add("limit", "20")
	q.Add("within", "100")
	q.Add("unit", "mile")
	u := "https://redsky.target.com/v3/stores/nearby/" + url.PathEscape(query) + "?" + q.Encode()
	data, err := GetRequest(u)
	if err != nil {
		return nil, errors.Wrap(err, "search stores")
	}
	var response []struct {
		Locations []*LocationInfo `json:"locations"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, errors.Wrap(err, "search stores")
	}
	if len(response) != 1 {
		return nil, errors.New("search stores: unexpected result count")
	}
	return response[0].Locations, nil
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
