package main

import (
	"encoding/json"
	"flag"
	"net/http"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	server := &Server{AssetDir: args.AssetDir}
	http.Handle("/", http.FileServer(http.Dir(args.AssetDir)))
	http.HandleFunc("/login", server.HandleLogin)
	http.ListenAndServe(args.Addr, nil)
}

type Server struct {
	AssetDir string
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: this.
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(obj)
}
