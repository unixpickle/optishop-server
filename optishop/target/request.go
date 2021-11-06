package target

import (
	"io/ioutil"
	"net/http"
)

// GetRequest performs a GET request on a Target URL,
// using specific necessary headers.
func GetRequest(url string) ([]byte, error) {
	resp, err := GetRequestRaw(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// GetRequestRaw performs a GET request on a Target URL,
// using specific necessary headers.
func GetRequestRaw(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-CLIENT-SLACK", "adaptive-ui-platform")
	return http.DefaultClient.Do(req)
}
