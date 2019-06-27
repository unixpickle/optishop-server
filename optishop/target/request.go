package target

import (
	"io/ioutil"
	"net/http"
)

// GetRequest performs a GET request on a Target URL,
// using specific necessary headers.
func GetRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-CLIENT-SLACK", "adaptive-ui-platform")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
