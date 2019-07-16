package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop/db"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	db, err := db.NewFileDB(args.DataDir)
	essentials.Must(err)

	server := &Server{
		AssetDir: args.AssetDir,
		DB:       db,
	}
	http.HandleFunc("/", server.HandleGeneral)
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/api/stores", AuthHandler(server.DB, server.HandleStoresAPI))
	http.ListenAndServe(args.Addr, nil)
}

type Server struct {
	AssetDir string
	DB       db.DB
}

func (s *Server) HandleGeneral(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		// Kind of a hack to make sure unauthenticated
		// homepage requests always get redirected to the
		// login page.
		AuthHandler(s.DB, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(s.AssetDir, "index.html"))
		})(w, r)
	} else {
		http.FileServer(http.Dir(s.AssetDir)).ServeHTTP(w, r)
	}
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, filepath.Join(s.AssetDir, "login.html"))
		return
	}

	userID, err := s.DB.Login(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	secret, err := s.DB.UserMetadata(userID, "secret")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: (url.Values{
			"user":   []string{string(userID)},
			"secret": []string{secret},
		}).Encode(),
		Expires: time.Now().Add(time.Hour * 24 * 30),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleStoresAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	stores, err := s.DB.Stores(user)
	if err != nil {
		serveError(w, r, err)
	} else {
		serveObject(w, r, stores)
	}
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	serveObject(w, r, obj)
}

func serveObject(w http.ResponseWriter, r *http.Request, obj interface{}) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(obj)
}
