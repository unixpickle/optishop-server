package main

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/unixpickle/ratelimit"
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
	return strings.HasPrefix(path.Clean(r.URL.Path), "/api")
}

// UncachedMux wraps a ServeMux to prevent caching on API
// endpoints and certain dynamic pages.
func UncachedMux(m *http.ServeMux) *http.ServeMux {
	result := http.NewServeMux()
	result.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if shouldPreventCaching(r) {
			// https://stackoverflow.com/questions/33880343/go-webserver-dont-cache-files-using-timestamp
			w.Header().Set("cache-control", "no-cache, private, max-age=0")
			w.Header().Set("pragma", "no-cache")
			w.Header().Set("expires", time.Unix(0, 0).Format(time.RFC1123))
		} else {
			w.Header().Set("expires", time.Now().Add(time.Minute).Format(http.TimeFormat))
		}
		h, _ := m.Handler(r)
		h.ServeHTTP(w, r)
	})
	return result
}

// RateLimitMux wraps a ServeMux to rate-limit requests
// for all APIs.
func RateLimitMux(s *Server, m *http.ServeMux) *http.ServeMux {
	heavyEndpoints := map[string]bool{
		"/login":  true,
		"/signup": true,
		"/route":  true,
	}
	limiter := ratelimit.NewTimeSliceLimiter(time.Minute*10, 250)
	namer := ratelimit.HTTPRemoteNamer{NumProxies: s.NumProxies}
	result := http.NewServeMux()
	result.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if IsAPIRequest(r) || heavyEndpoints[path.Clean(r.URL.Path)] {
			if limiter.Limit(namer.Name(r)) {
				s.ServeError(w, r, errors.New("You have made too many requests. "+
					"Try again in 10 minutes"))
				return
			}
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
		"/logout": true,
		"/route":  true,
		"/signup": true,
	}
	return IsAPIRequest(r) || pages[r.URL.Path]
}
