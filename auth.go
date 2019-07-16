package main

import (
	"context"
	"net/http"
	"net/url"

	"github.com/unixpickle/optishop-server/optishop/db"
)

type UserKeyType int

// UserKey is the context key used to store a db.UserID.
var UserKey UserKeyType

// AuthHandler wraps an HTTP handler to ensure that the
// handler is only called for authenticated requests.
//
// The handler will get a UserKey added to its request
// context.
func AuthHandler(d db.DB, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := checkAuth(d, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
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
