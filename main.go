package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
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
	http.Handle("/", http.FileServer(http.Dir(args.AssetDir)))
	http.HandleFunc("/login/submit", server.HandleLogin)
	http.ListenAndServe(args.Addr, nil)
}

type Server struct {
	AssetDir string
	DB       db.DB
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
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

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(obj)
}
