package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/unixpickle/optishop-server/optishop/db"
)

type UserKeyType int

// UserKey is the context key used to store a db.UserID.
var UserKey UserKeyType

// SetAuthCookie sets a user cookie for a request.
func SetAuthCookie(w http.ResponseWriter, user db.UserID, secret string) {
	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: (url.Values{
			"user":   []string{string(user)},
			"secret": []string{secret},
		}).Encode(),
		Expires: time.Now().Add(time.Hour * 24 * 30),
	})
}

// AuthHandler wraps an HTTP handler to ensure that the
// handler is only called for authenticated requests.
//
// The handler will get a UserKey added to its request
// context.
func AuthHandler(d db.DB, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := checkAuth(d, r)
		if !ok {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				serveError(w, r, errors.New("not authenticated"))
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
		} else {
			f(w, r.WithContext(context.WithValue(r.Context(), UserKey, user)))
		}
	}
}

func checkAuth(d db.DB, r *http.Request) (db.UserID, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	values, err := url.ParseQuery(cookie.Value)
	if err != nil {
		return "", false
	}
	user := values.Get("user")
	secret := values.Get("secret")
	realSecret, err := d.UserMetadata(db.UserID(user), "secret")
	if err != nil || realSecret != secret {
		return "", false
	}
	return db.UserID(user), true
}
