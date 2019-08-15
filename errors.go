package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

var errorMap = map[string]string{
	"get store: create store: get map info: store does not have a map": "There is no digital map available for this store. Try choosing a different store.",
	"could not locate product within store":                            "This product has no listed aisle or department section. Try choosing a different, similar product.",
	"check login: password incorrect":                                  "The password you entered is incorrect.",
	"check login: user does not exist":                                 "The username you entered does not exist.",
	"create user: user already exists":                                 "That username is already in use.",
	"passwords do not match":                                           "The passwords you entered do not match",
	"not authenticated":                                                "You are no longer signed in. Please refresh the page and sign in.",
	"get store: store not found":                                       "The store could not be found. Did you delete it?",
	"remove list entry: entry not found":                               "The entry does not exist. Did you delete it?",
	"the product cannot be specifically located":                       "The product's exact location is unknown.",
}

var errorRegexes = map[*regexp.Regexp]string{
	regexp.MustCompile("^locate product: aisle (.*) is missing from the map$"): "The product is located at aisle $1, but $1 is missing from the map.",
}

// HumanizeError turns an error message into a more
// user-friendly message.
func HumanizeError(err error) error {
	msg := err.Error()
	if str, ok := errorMap[msg]; ok {
		return errors.New(str)
	}
	for expr, replacement := range errorRegexes {
		if expr.MatchString(msg) {
			return errors.New(expr.ReplaceAllString(msg, replacement))
		}
	}
	return err
}

// ServeFormError redirects the user to an error page when
// a form POST results in an error.
func ServeFormError(w http.ResponseWriter, r *http.Request, err error) {
	path := r.URL.Path
	http.Redirect(w, r, path+"?error="+url.QueryEscape(HumanizeError(err).Error()),
		http.StatusSeeOther)
	LogRequest(r, "serving form error: %s", err.Error())
}

// ServeError serves errors for API and page requests.
//
// The only situation in which ServeError should not be
// used is when the error occurs on a login or signup form
// due to some credential issue, in which case
// ServeFormError should be used instead.
func (s *Server) ServeError(w http.ResponseWriter, r *http.Request, err error) {
	message := HumanizeError(err).Error()
	LogRequest(r, "serving error: %s", message)

	if IsAPIRequest(r) {
		obj := map[string]string{"error": message}
		ServeObject(w, r, obj)
		return
	}

	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "error.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData = bytes.Replace(pageData, []byte("INSERT_ERROR_HERE"), []byte(message), 1)
	w.Write(pageData)
}
