package target

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

type Client struct {
	preloadedState map[string]interface{}
	visitorID      string
}

func NewClient() (*Client, error) {
	rawResp, err := GetRequestRaw("https://www.target.com")
	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()
	resp, err := ioutil.ReadAll(rawResp.Body)
	if err != nil {
		return nil, err
	}

	preloadMarker := []byte("window.__PRELOADED_STATE__= ")
	idx := bytes.Index(resp, preloadMarker)
	jsonPayload := resp[idx+len(preloadMarker):]
	jsonPayload = bytes.ReplaceAll(jsonPayload, []byte("undefined"), []byte("null"))
	jsonPayload = bytes.ReplaceAll(jsonPayload, []byte("new Set([])"), []byte("[]"))
	var state map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(jsonPayload)).Decode(&state); err != nil {
		return nil, errors.Wrap(err, "new store source")
	}
	for _, cookie := range rawResp.Cookies() {
		if cookie.Name == "visitorId" {
			return &Client{
				preloadedState: state,
				visitorID:      cookie.Value,
			}, nil
		}
	}
	return nil, errors.New("new store source: no visitorId cookie set")
}

// Key gets the key field for various requests.
func (c *Client) Key() string {
	config, ok := c.preloadedState["config"].(map[string]interface{})
	if !ok {
		return ""
	}
	firefly, ok := config["firefly"].(map[string]interface{})
	if !ok {
		return ""
	}
	data, _ := firefly["apiKey"].(string)
	return data
}

// VisitorID gets the visitorId cookie.
func (c *Client) VisitorID() string {
	return c.visitorID
}
