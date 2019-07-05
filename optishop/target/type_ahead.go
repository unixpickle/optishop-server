package target

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// A Suggestion is a result from a type-ahead prediction.
type Suggestion struct {
	SuggestionType string        `json:"suggestionType"`
	Label          string        `json:"label"`
	Location       string        `json:"location"`
	Name           string        `json:"name"`
	SubResults     []*Suggestion `json:"subResults"`
}

// TypeAheadResults contains the results of a type-ahead
// prediction.
type TypeAheadResults struct {
	Suggestions []*Suggestion    `json:"suggestions"`
	Metadata    *RequestMetadata `json:"metaData"`
}

// TypeAhead gets auto-complete suggestions for a given
// textual query.
func TypeAhead(query string) (*TypeAheadResults, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("ctgryVal", "0|ALL|matchallpartial|all categories")
	u := "https://typeahead.target.com/autocomplete/TypeAheadSearch/v2" + q.Encode()
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrap(err, "type ahead")
	}
	defer resp.Body.Close()
	var res TypeAheadResults
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "type ahead")
	}
	return &res, nil
}
