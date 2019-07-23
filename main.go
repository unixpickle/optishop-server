package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/ajstarks/svgo/float"
	"github.com/unixpickle/optishop-server/optishop/visualize"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/db"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	db, err := db.NewFileDB(args.DataDir)
	essentials.Must(err)

	sources, err := LoadStoreSources()
	essentials.Must(err)

	server := &Server{
		AssetDir:   args.AssetDir,
		DB:         db,
		Sources:    sources,
		StoreCache: NewStoreCache(sources),
	}
	http.HandleFunc("/", server.HandleGeneral)
	http.HandleFunc("/list", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleList)))
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/route", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleRoute)))
	http.HandleFunc("/signup", server.HandleSignup)
	http.HandleFunc("/api/additem", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleAddItemAPI)))
	http.HandleFunc("/api/addstore", AuthHandler(server.DB, server.HandleAddStoreAPI))
	http.HandleFunc("/api/chpass", AuthHandler(server.DB, server.HandleChpassAPI))
	http.HandleFunc("/api/inventoryquery", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleInventoryQueryAPI)))
	http.HandleFunc("/api/list", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleListAPI)))
	http.HandleFunc("/api/removeitem", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleRemoveItemAPI)))
	http.HandleFunc("/api/removestore", AuthHandler(server.DB, server.HandleRemoveStoreAPI))
	http.HandleFunc("/api/sort", AuthHandler(server.DB,
		StoreHandler(server.DB, server.StoreCache, server.HandleSortAPI)))
	http.HandleFunc("/api/storequery", AuthHandler(server.DB, server.HandleStoreQueryAPI))
	http.HandleFunc("/api/stores", AuthHandler(server.DB, server.HandleStoresAPI))
	http.ListenAndServe(args.Addr, nil)
}

type Server struct {
	AssetDir   string
	DB         db.DB
	Sources    map[string]optishop.StoreSource
	StoreCache *StoreCache
}

func (s *Server) HandleGeneral(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		// Hack for the stores page, since it always gets
		// clumped in with the other asset requests.
		AuthHandler(s.DB, s.HandleStores)(w, r)
	} else {
		http.FileServer(http.Dir(s.AssetDir)).ServeHTTP(w, r)
	}
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "list.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	record, err := s.DB.Store(userID, storeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	storeDesc := &ClientStoreDesc{
		ID:      string(record.ID),
		Source:  record.Info.SourceName,
		Name:    record.Info.StoreName,
		Address: record.Info.StoreAddress,
	}

	storeData, err := json.Marshal(storeDesc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := s.getClientListItems(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	listData, err := json.Marshal(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData = bytes.Replace(pageData, []byte("INSERT_DATA_HERE"), listData, 1)
	pageData = bytes.Replace(pageData, []byte("INSERT_STORE_DATA_HERE"), storeData, 1)
	w.Write(pageData)
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
	secret, err := s.DB.UserMetadata(userID, SecretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	SetAuthCookie(w, userID, secret)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleRoute(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "route.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	entries, err := s.DB.ListEntries(userID, storeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	paths, err := RoutePaths(entries, store, optishop.NewFloorConnectorCached(store.Layout()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var imageData bytes.Buffer
	canvas := svg.New(&imageData)

	width, height, _ := visualize.MultiFloorGeometry(store.Layout())
	canvas.Start(width, height, fmt.Sprintf("viewBox=\"0 0 %f %f\"", width, height))

	visualize.DrawFloors(canvas, store.Layout())
	for _, path := range paths {
		visualize.DrawFloorPath(canvas, store.Layout(), path)
	}

	canvas.End()

	// Remove the width/height attributes so that the SVG
	// has a dynamic size.
	expr := regexp.MustCompile(`<svg width="[0-9\.]*" height="[0-9\.]*"`)
	data := expr.ReplaceAll(imageData.Bytes(), []byte("<svg"))

	page := bytes.Replace(pageData, []byte("INSERT_IMAGE_HERE"), data, 1)
	w.Write(page)
}

func (s *Server) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, filepath.Join(s.AssetDir, "signup.html"))
		return
	}

	if r.FormValue("password") != r.FormValue("confirm") {
		http.Redirect(w, r, "/signup?error="+url.QueryEscape("passwords do not match"),
			http.StatusSeeOther)
		return
	}

	secret, err := GenerateSecret()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	signatureKey, err := GenerateSecret()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	metadata := map[string]string{
		SecretKey:    secret,
		SignatureKey: signatureKey,
	}

	userID, err := s.DB.CreateUser(r.FormValue("username"), r.FormValue("password"), metadata)
	if err != nil {
		http.Redirect(w, r, "/signup?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	SetAuthCookie(w, userID, secret)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleStores(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "stores.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clientStores, err := s.getClientStores(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	storeData, err := json.Marshal(clientStores)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := bytes.Replace(pageData, []byte("INSERT_DATA_HERE"), storeData, 1)
	w.Write(page)
}

func (s *Server) HandleAddItemAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	var data []byte
	if err := json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
		serveError(w, r, err)
		return
	}

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		serveError(w, r, err)
		return
	}

	if SignInventoryItem(sigKey, storeID, data) != r.FormValue("signature") {
		serveError(w, r, errors.New("invalid signature"))
		return
	}

	product, err := store.UnmarshalProduct(data)
	if err != nil {
		serveError(w, r, err)
		return
	}
	zone, err := store.Locate(product)
	if err != nil {
		serveError(w, r, err)
		return
	}
	if zone == nil {
		serveError(w, r, errors.New("could not locate product within store"))
		return
	}
	floor := store.Layout().ZoneFloor(zone)

	_, err = s.DB.AddListEntry(user, storeID, &db.ListEntryInfo{
		InventoryProductData: data,
		Zone:                 zone,
		Floor:                floor,
	})
	if err != nil {
		serveError(w, r, err)
		return
	}

	s.HandleListAPI(w, r)
}

func (s *Server) HandleAddStoreAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)

	sourceName := r.FormValue("source")
	signature := r.FormValue("signature")

	var data []byte
	if err := json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
		serveError(w, r, err)
		return
	}

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		serveError(w, r, err)
		return
	}

	if SignStore(sigKey, sourceName, data) != signature {
		serveError(w, r, errors.New("invalid signature"))
		return
	}

	source, ok := s.Sources[sourceName]
	if !ok {
		serveError(w, r, errors.New("missing store source"))
		return
	}

	store, err := source.UnmarshalStoreDesc(data)
	if err != nil {
		serveError(w, r, err)
		return
	}

	storeID, err := s.DB.AddStore(user, &db.StoreInfo{
		SourceName:   sourceName,
		StoreName:    store.Name(),
		StoreAddress: store.Address(),
		StoreData:    data,
	})
	if err != nil {
		serveError(w, r, err)
		return
	}
	serveObject(w, r, storeID)
}

func (s *Server) HandleChpassAPI(w http.ResponseWriter, r *http.Request) {
	secret, err := GenerateSecret()
	if err != nil {
		serveError(w, r, err)
		return
	}
	user := r.Context().Value(UserKey).(db.UserID)
	old := r.FormValue("old")
	new := r.FormValue("new")
	if err := s.DB.Chpass(user, old, new); err != nil {
		serveError(w, r, err)
		return
	}
	if err := s.DB.SetUserMetadata(user, SecretKey, secret); err != nil {
		serveError(w, r, errors.New("failed to log out other sessions"))
		return
	}
	SetAuthCookie(w, user, secret)
	serveObject(w, r, map[string]string{})
}

func (s *Server) HandleInventoryQueryAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		serveError(w, r, err)
		return
	}

	rawResults, err := store.Search(r.FormValue("query"))
	if err != nil {
		serveError(w, r, err)
		return
	}

	results := []*ClientInventoryItem{}
	for _, result := range rawResults {
		data, err := store.MarshalProduct(result)
		if err != nil {
			serveError(w, r, err)
			return
		}
		results = append(results, &ClientInventoryItem{
			ClientListItem: NewClientListItem(result),
			Data:           data,
			Signature:      SignInventoryItem(sigKey, storeID, data),
		})
	}

	serveObject(w, r, results)
}

func (s *Server) HandleListAPI(w http.ResponseWriter, r *http.Request) {
	items, err := s.getClientListItems(r)
	if err != nil {
		serveError(w, r, err)
		return
	}
	serveObject(w, r, items)
}

func (s *Server) getClientListItems(r *http.Request) ([]*ClientListItem, error) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	listEntries, err := s.DB.ListEntries(user, storeID)
	if err != nil {
		return nil, err
	}

	results := []*ClientListItem{}
	for _, entry := range listEntries {
		invProd, err := store.UnmarshalProduct(entry.Info.InventoryProductData)
		if err != nil {
			return nil, err
		}
		item := NewClientListItem(invProd)
		item.ID = string(entry.ID)
		if entry.Info.Zone != nil {
			item.ZoneName = entry.Info.Zone.Name
		}
		results = append(results, item)
	}

	return results, nil
}

func (s *Server) HandleRemoveItemAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	store := db.StoreID(r.FormValue("store"))
	item := db.ListEntryID(r.FormValue("item"))
	if err := s.DB.RemoveListEntry(user, store, item); err != nil {
		serveError(w, r, err)
		return
	}
	s.HandleListAPI(w, r)
}

func (s *Server) HandleRemoveStoreAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	store := db.StoreID(r.FormValue("store"))
	if err := s.DB.RemoveStore(user, store); err != nil {
		serveError(w, r, err)
		return
	}
	s.HandleStoresAPI(w, r)
}

func (s *Server) HandleSortAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	list, err := s.DB.ListEntries(user, storeID)
	if err != nil {
		serveError(w, r, err)
		return
	}

	entries, err := SortEntries(list, store, optishop.NewFloorConnectorCached(store.Layout()))
	if err != nil {
		serveError(w, r, err)
		return
	}

	newIDs := make([]db.ListEntryID, len(entries))
	for i, entry := range entries {
		newIDs[i] = entry.ID
	}
	if err := s.DB.PermuteListEntries(user, storeID, newIDs); err != nil {
		serveError(w, r, err)
		return
	}

	s.HandleListAPI(w, r)
}

func (s *Server) HandleStoreQueryAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		serveError(w, r, err)
		return
	}

	query := r.FormValue("query")

	responses := []*ClientStoreDesc{}
	for name, source := range s.Sources {
		results, err := source.QueryStores(query)
		if err != nil {
			serveError(w, r, err)
			return
		}
		for _, result := range results {
			data, err := source.MarshalStoreDesc(result)
			if err != nil {
				serveError(w, r, err)
				return
			}
			responses = append(responses, &ClientStoreDesc{
				Source:  name,
				Name:    result.Name(),
				Address: result.Address(),
				Data:    data,

				Signature: SignStore(sigKey, name, data),
			})
		}
	}

	serveObject(w, r, responses)
}

func (s *Server) HandleStoresAPI(w http.ResponseWriter, r *http.Request) {
	clientStores, err := s.getClientStores(r)
	if err != nil {
		serveError(w, r, err)
		return
	}
	serveObject(w, r, clientStores)
}

func (s *Server) getClientStores(r *http.Request) ([]*ClientStoreDesc, error) {
	user := r.Context().Value(UserKey).(db.UserID)

	stores, err := s.DB.Stores(user)
	if err != nil {
		return nil, err
	}

	clientStores := []*ClientStoreDesc{}
	for _, record := range stores {
		clientStores = append(clientStores, &ClientStoreDesc{
			ID:      string(record.ID),
			Source:  record.Info.SourceName,
			Name:    record.Info.StoreName,
			Address: record.Info.StoreAddress,
		})
	}

	return clientStores, nil
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	serveObject(w, r, obj)
}

func serveObject(w http.ResponseWriter, r *http.Request, obj interface{}) {
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(obj)
}
