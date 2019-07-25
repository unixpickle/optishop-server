package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ServeObject responds to an API request with a JSON
// serialized object.
func ServeObject(w http.ResponseWriter, r *http.Request, obj interface{}) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(obj)
}

// IsAPIRequest checks if a request is an API request
// (versus a user-facing page).
func IsAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api")
}

// UncachedMux wraps a ServeMux to prevent caching on API
// endpoints and certain dynamic pages.
func UncachedMux(m *http.ServeMux) *http.ServeMux {
	result := http.NewServeMux()
	result.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if shouldPreventCaching(r) {
			// https://stackoverflow.com/questions/33880343/go-webserver-dont-cache-files-using-timestamp
			r.Header.Set("cache-control", "no-cache, private, max-age=0")
			r.Header.Set("pragma", "no-cache")
			r.Header.Set("expires", time.Unix(0, 0).Format(time.RFC1123))
		}
		h, _ := m.Handler(r)
		h.ServeHTTP(w, r)
	})
	return result
}

func shouldPreventCaching(r *http.Request) bool {
	pages := map[string]bool{
		"/":       true,
		"/list":   true,
		"/login":  true,
		"/route":  true,
		"/signup": true,
	}
	return IsAPIRequest(r) || pages[r.URL.Path]
}
