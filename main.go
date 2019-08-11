package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

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
		NumProxies: args.NumProxies,
		DB:         db,
		Sources:    sources,
		StoreCache: NewStoreCache(sources),
	}
	http.HandleFunc("/", server.HandleGeneral)
	http.HandleFunc("/list", server.AuthHandler(server.StoreHandler(server.HandleList)))
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/logout", server.HandleLogout)
	http.HandleFunc("/route", server.AuthHandler(server.StoreHandler(server.HandleRoute)))
	http.HandleFunc("/signup", server.HandleSignup)
	http.HandleFunc("/api/additem",
		server.AuthHandler(server.StoreHandler(server.HandleAddItemAPI)))
	http.HandleFunc("/api/addstore", server.AuthHandler(server.HandleAddStoreAPI))
	http.HandleFunc("/api/chpass", server.AuthHandler(server.HandleChpassAPI))
	http.HandleFunc("/api/inventoryquery",
		server.AuthHandler(server.StoreHandler(server.HandleInventoryQueryAPI)))
	http.HandleFunc("/api/list", server.AuthHandler(server.StoreHandler(server.HandleListAPI)))
	http.HandleFunc("/api/removeitem",
		server.AuthHandler(server.StoreHandler(server.HandleRemoveItemAPI)))
	http.HandleFunc("/api/removestore", server.AuthHandler(server.HandleRemoveStoreAPI))
	http.HandleFunc("/api/sort", server.AuthHandler(server.StoreHandler(server.HandleSortAPI)))
	http.HandleFunc("/api/storequery", server.AuthHandler(server.HandleStoreQueryAPI))
	http.HandleFunc("/api/stores", server.AuthHandler(server.HandleStoresAPI))
	http.ListenAndServe(args.Addr, RateLimitMux(server, UncachedMux(http.DefaultServeMux)))
}

type Server struct {
	AssetDir   string
	NumProxies int

	DB         db.DB
	Sources    map[string]optishop.StoreSource
	StoreCache *StoreCache
}

func (s *Server) HandleGeneral(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		// Hack for the stores page, since it always gets
		// clumped in with the other asset requests.
		s.AuthHandler(s.HandleStores)(w, r)
	} else {
		http.FileServer(http.Dir(s.AssetDir)).ServeHTTP(w, r)
	}
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "list.html"))
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	userID := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	record, err := s.DB.Store(userID, storeID)
	if err != nil {
		s.ServeError(w, r, err)
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
		s.ServeError(w, r, err)
		return
	}

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
		ServeFormError(w, r, err)
		return
	}
	secret, err := s.DB.UserMetadata(userID, SecretKey)
	if err != nil {
		s.ServeError(w, r, err)
	}
	SetAuthCookie(w, userID, secret)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(time.Second),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) HandleRoute(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "route.html"))
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	userID := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	entries, err := s.DB.ListEntries(userID, storeID)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	connector := optishop.NewFloorConnectorCached(store.Layout())
	paths, sorted, err := RoutePaths(entries, store, connector)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	var imageData bytes.Buffer
	canvas := svg.New(&imageData)

	width, height, _ := visualize.MultiFloorGeometry(store.Layout())
	canvas.Start(width, height, fmt.Sprintf("viewBox=\"0 0 %f %f\"", width, height))

	visualize.MultiFloorLoop(store.Layout(), func(f *optishop.Floor, x, y float64) {
		visualize.DrawFloorPolygons(canvas, f, x, y)
	})
	for _, path := range paths {
		visualize.DrawFloorPath(canvas, store.Layout(), path)
	}
	fontSize := width * visualize.FontSizeFrac * 2
	visualize.MultiFloorLoop(store.Layout(), func(f *optishop.Floor, x, y float64) {
		floorIdx := store.Layout().FloorIndex(f)
		var zones []*optishop.Zone
		for _, entry := range sorted {
			if store.Layout().ZoneFloor(entry.Info.Zone) == floorIdx {
				zones = append(zones, entry.Info.Zone)
			}
		}
		visualize.DrawZoneLabels(canvas, zones, x, y, fontSize)
	})

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
		ServeFormError(w, r, errors.New("passwords do not match"))
		return
	}

	secret, err := GenerateSecret()
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	signatureKey, err := GenerateSecret()
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	metadata := map[string]string{
		SecretKey:    secret,
		SignatureKey: signatureKey,
	}

	userID, err := s.DB.CreateUser(r.FormValue("username"), r.FormValue("password"), metadata)
	if err != nil {
		ServeFormError(w, r, err)
		return
	}

	SetAuthCookie(w, userID, secret)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleStores(w http.ResponseWriter, r *http.Request) {
	pageData, err := ioutil.ReadFile(filepath.Join(s.AssetDir, "stores.html"))
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	userID := r.Context().Value(UserKey).(db.UserID)
	username, err := s.DB.Username(userID)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	usernameData, err := json.Marshal(username)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	pageData = bytes.Replace(pageData, []byte("INSERT_USERNAME"), usernameData, 1)
	w.Write(pageData)
}

func (s *Server) HandleAddItemAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)
	force := r.FormValue("force") == "1"

	var data []byte
	if err := json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
		s.ServeError(w, r, err)
		return
	}

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	if SignInventoryItem(sigKey, storeID, data) != r.FormValue("signature") {
		s.ServeError(w, r, errors.New("invalid signature"))
		return
	}

	product, err := store.UnmarshalProduct(data)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	zone, err := store.Locate(product)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	if zone == nil {
		s.ServeError(w, r, errors.New("could not locate product within store"))
		return
	}
	if !zone.Specific && !force {
		s.ServeError(w, r, errors.New("the product cannot be specifically located"))
		return
	}
	floor := store.Layout().ZoneFloor(zone)

	_, err = s.DB.AddListEntry(user, storeID, &db.ListEntryInfo{
		InventoryProductData: data,
		Zone:                 zone,
		Floor:                floor,
	})
	if err != nil {
		s.ServeError(w, r, err)
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
		s.ServeError(w, r, err)
		return
	}

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	if SignStore(sigKey, sourceName, data) != signature {
		s.ServeError(w, r, errors.New("invalid signature"))
		return
	}

	// Make sure that the store has an available map, etc.
	if _, err := s.StoreCache.GetStore(sourceName, data); err != nil {
		s.ServeError(w, r, err)
		return
	}

	source, ok := s.Sources[sourceName]
	if !ok {
		s.ServeError(w, r, errors.New("missing store source"))
		return
	}

	store, err := source.UnmarshalStoreDesc(data)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	storeID, err := s.DB.AddStore(user, &db.StoreInfo{
		SourceName:   sourceName,
		StoreName:    store.Name(),
		StoreAddress: store.Address(),
		StoreData:    data,
	})
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	ServeObject(w, r, storeID)
}

func (s *Server) HandleChpassAPI(w http.ResponseWriter, r *http.Request) {
	secret, err := GenerateSecret()
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	user := r.Context().Value(UserKey).(db.UserID)
	old := r.FormValue("old")
	new := r.FormValue("new")
	if err := s.DB.Chpass(user, old, new); err != nil {
		s.ServeError(w, r, err)
		return
	}
	if err := s.DB.SetUserMetadata(user, SecretKey, secret); err != nil {
		s.ServeError(w, r, errors.New("failed to log out other sessions"))
		return
	}
	SetAuthCookie(w, user, secret)
	ServeObject(w, r, map[string]string{})
}

func (s *Server) HandleInventoryQueryAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := r.Context().Value(StoreIDKey).(db.StoreID)
	store := r.Context().Value(StoreKey).(optishop.Store)

	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	rawResults, err := store.Search(r.FormValue("query"))
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	results := []*ClientInventoryItem{}
	for _, result := range rawResults {
		data, err := store.MarshalProduct(result)
		if err != nil {
			s.ServeError(w, r, err)
			return
		}
		results = append(results, &ClientInventoryItem{
			ClientListItem: NewClientListItem(result),
			Data:           data,
			Signature:      SignInventoryItem(sigKey, storeID, data),
		})
	}

	ServeObject(w, r, results)
}

func (s *Server) HandleListAPI(w http.ResponseWriter, r *http.Request) {
	items, err := s.getClientListItems(r)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	ServeObject(w, r, items)
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
		s.ServeError(w, r, err)
		return
	}
	s.HandleListAPI(w, r)
}

func (s *Server) HandleRemoveStoreAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	store := db.StoreID(r.FormValue("store"))
	if err := s.DB.RemoveStore(user, store); err != nil {
		s.ServeError(w, r, err)
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
		s.ServeError(w, r, err)
		return
	}

	entries, err := SortEntries(list, store, optishop.NewFloorConnectorCached(store.Layout()))
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	newIDs := make([]db.ListEntryID, len(entries))
	for i, entry := range entries {
		newIDs[i] = entry.ID
	}
	if err := s.DB.PermuteListEntries(user, storeID, newIDs); err != nil {
		s.ServeError(w, r, err)
		return
	}

	s.HandleListAPI(w, r)
}

func (s *Server) HandleStoreQueryAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}

	query := r.FormValue("query")

	responses := []*ClientStoreDesc{}
	for name, source := range s.Sources {
		results, err := source.QueryStores(query)
		if err != nil {
			s.ServeError(w, r, err)
			return
		}
		for _, result := range results {
			data, err := source.MarshalStoreDesc(result)
			if err != nil {
				s.ServeError(w, r, err)
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

	ServeObject(w, r, responses)
}

func (s *Server) HandleStoresAPI(w http.ResponseWriter, r *http.Request) {
	clientStores, err := s.getClientStores(r)
	if err != nil {
		s.ServeError(w, r, err)
		return
	}
	ServeObject(w, r, clientStores)
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
