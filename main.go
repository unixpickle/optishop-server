package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/target"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	store, err := target.NewStore(args.StoreID)
	essentials.Must(err)

	server := &Server{Store: store, AssetDir: args.AssetDir}
	http.HandleFunc("/", server.HandleRoot)
	http.HandleFunc("/list", server.HandleList)
	http.HandleFunc("/search", server.HandleSearch)
	http.HandleFunc("/add", server.HandleAdd)
	http.HandleFunc("/delete", server.HandleDelete)
}

type Server struct {
	Store    optishop.Store
	AssetDir string

	List  []optishop.InventoryProduct
	Zones []*optishop.Zone
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.AssetDir, "index.html"))
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	list := ConvertListItems(s.List)
	resultObj := map[string]interface{}{"list": list, "zones": s.Zones}
	json.NewEncoder(w).Encode(resultObj)
}

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	results, err := s.Store.Search(r.FormValue("query"))
	if err != nil {
		serveError(w, r, err)
		return
	}
	exactResults := make([][]byte, len(results))
	for i, res := range results {
		exactResults[i], err = s.Store.MarshalProduct(res)
		if err != nil {
			serveError(w, r, err)
			return
		}
	}
	resultObj := map[string]interface{}{
		"results":      ConvertListItems(results),
		"exactResults": exactResults,
	}
	json.NewEncoder(w).Encode(resultObj)
}

func (s *Server) HandleAdd(w http.ResponseWriter, r *http.Request) {
	rawItemData := r.FormValue("item")
	var itemData []byte
	if err := json.Unmarshal([]byte(rawItemData), &itemData); err != nil {
		serveError(w, r, err)
		return
	}
	product, err := s.Store.UnmarshalProduct(itemData)
	if err != nil {
		serveError(w, r, err)
		return
	}
	zone, err := s.Store.Locate(product)
	if err != nil {
		serveError(w, r, err)
		return
	}
	s.List = append(s.List, product)
	s.Zones = append(s.Zones, zone)

	s.HandleList(w, r)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	for i, prod := range s.List {
		if prod.Name() == name {
			essentials.OrderedDelete(&s.List, i)
			essentials.OrderedDelete(&s.Zones, i)
			s.HandleList(w, r)
			return
		}
	}
	serveError(w, r, errors.New("no list item found"))
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(obj)
}
