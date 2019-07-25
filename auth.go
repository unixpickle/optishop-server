package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop/db"
)

type UserKeyType int

// UserKey is the context key used to store a db.UserID.
var UserKey UserKeyType

// SecretKey is the user metadata field used to store the
// user secret in the database.
const SecretKey = "secret"

// GenerateSecret generates a random string which is
// cryptographically unpredictable.
func GenerateSecret() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", errors.Wrap(err, "generate secret")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

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
func (s *Server) AuthHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := checkAuth(s.DB, r)
		if !ok {
			if IsAPIRequest(r) {
				s.ServeError(w, r, errors.New("not authenticated"))
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}
		f(w, r.WithContext(context.WithValue(r.Context(), UserKey, user)))
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
