package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"path/filepath"

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
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/signup", server.HandleSignup)
	http.HandleFunc("/api/addstore", AuthHandler(server.DB, server.HandleAddStoreAPI))
	http.HandleFunc("/api/chpass", AuthHandler(server.DB, server.HandleChpassAPI))
	http.HandleFunc("/api/list", AuthHandler(server.DB, server.HandleListAPI))
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
	secret, err := s.DB.UserMetadata(userID, SecretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	SetAuthCookie(w, userID, secret)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, filepath.Join(s.AssetDir, "signup.html"))
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
		SourceName: sourceName,
		StoreName:  store.Name(),
		StoreData:  data,
	})
	if err != nil {
		serveError(w, r, err)
	} else {
		serveObject(w, r, storeID)
	}
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

func (s *Server) HandleListAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	storeID := db.StoreID(r.FormValue("store"))

	storeRecord, err := s.DB.Store(user, storeID)
	if err != nil {
		serveError(w, r, err)
		return
	}

	store, err := s.StoreCache.GetStore(storeRecord.Info.SourceName, storeRecord.Info.StoreData)
	if err != nil {
		serveError(w, r, err)
		return
	}

	listEntries, err := s.DB.ListEntries(user, storeID)
	if err != nil {
		serveError(w, r, err)
		return
	}

	var results []*ListItem
	for _, entry := range listEntries {
		invProd, err := store.UnmarshalProduct(entry.Info.InventoryProductData)
		if err != nil {
			serveError(w, r, err)
			return
		}
		results = append(results, &ListItem{
			Name:        invProd.Name(),
			PhotoURL:    invProd.PhotoURL(),
			Description: invProd.Description(),
			InStock:     invProd.InStock(),
			Price:       invProd.Price(),
		})
	}

	serveObject(w, r, results)
}

func (s *Server) HandleStoreQueryAPI(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(db.UserID)
	sigKey, err := s.DB.UserMetadata(user, SignatureKey)
	if err != nil {
		serveError(w, r, err)
	}

	query := r.FormValue("query")

	var responses []*StoreDesc
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
			responses = append(responses, &StoreDesc{
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
