package target

import (
	"bytes"
	"encoding/json"
)

type Client struct {
	preloadedState map[string]interface{}
}

func NewClient() (*Client, error) {
	resp, err := GetRequest("https://www.target.com")
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
		return nil, err
	}
	return &Client{preloadedState: state}, nil
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
