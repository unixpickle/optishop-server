package target

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// ProductDepartment parses the department name from a
// product's buy URL.
func ProductDepartment(buyURL string) (string, error) {
	var response struct {
		Graph []struct {
			Type string `json:"@type"`
			List []struct {
				Type     string `json:"@type"`
				Position int    `json:"position"`
				Item     struct {
					Name string `json:"name"`
				} `json:"item"`
			} `json:"itemListElement"`
		} `json:"@graph"`
	}
	if err := ProductDetails(buyURL, &response); err != nil {
		return "", errors.Wrap(err, "product department")
	}
	for _, elem := range response.Graph {
		if elem.Type == "BreadcrumbList" {
			for _, bc := range elem.List {
				if bc.Position == 1 {
					return bc.Item.Name, nil
				}
			}
			return "", errors.New("product department: could not find breadcrumb")
		}
	}
	return "", errors.New("product department: could not find breadcrumbs")
}

// ProductDetails parses the metadata JSON blob on a
// product page.
func ProductDetails(buyURL string, out interface{}) error {
	data, err := GetRequest(buyURL)
	if err != nil {
		return errors.Wrap(err, "product details")
	}
	page := string(data)
	starter := `<script class="" type="application/ld+json">`
	idx := strings.Index(page, starter)
	if idx < 0 {
		return errors.New("product details: no data found")
	}
	page = page[idx+len(starter):]
	dec := json.NewDecoder(bytes.NewReader([]byte(page)))
	if err := dec.Decode(&out); err != nil {
		return errors.Wrap(err, "product details")
	}
	return nil
}
