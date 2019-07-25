package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

var errorMap = map[string]string{
	"get store: create store: get map info: store does not have a map": "There is not digital map available for this store. Try choosing a different store.",
	"could not locate product within store":                            "This product has no listed aisle or department section. Try choosing a different, similar product.",
	"check login: password incorrect":                                  "The password you entered is incorrect.",
	"check login: user does not exist":                                 "The username you entered does not exist.",
	"create user: user already exists":                                 "That username is already in use.",
	"passwords do not match":                                           "The passwords you entered do not match",
	"not authenticated":                                                "You are no longer signed in. Please refresh the page and sign in.",
}

// HumanizeError turns an error message into a more
// user-friendly message.
func HumanizeError(err error) error {
	if str, ok := errorMap[err.Error()]; ok {
		return errors.New(str)
	}
	return err
}

// ServeFormError redirects the user to an error page when
// a form POST results in an error.
func ServeFormError(w http.ResponseWriter, r *http.Request, err error) {
	path := r.URL.Path
	http.Redirect(w, r, path+"?error="+url.QueryEscape(HumanizeError(err).Error()),
		http.StatusSeeOther)
}

// ServeError serves errors for API and page requests.
//
// The only situation in which ServeError should not be
// used is when the error occurs on a login or signup form
// due to some credential issue, in which case
// ServeFormError should be used instead.
func ServeError(w http.ResponseWriter, r *http.Request, err error) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		obj := map[string]string{"error": HumanizeError(err).Error()}
		ServeObject(w, r, obj)
	} else {
		http.Error(w, HumanizeError(err).Error(), http.StatusInternalServerError)
	}
}
