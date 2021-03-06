package main

import (
	"flag"
	"net/http"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop/db"
	"github.com/unixpickle/optishop-server/serverapi"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	var dbInstance db.DB
	var err error
	if args.LocalMode {
		dbInstance, err = db.NewLocalDB(args.DataDir)
	} else {
		dbInstance, err = db.NewFileDB(args.DataDir)
	}
	essentials.Must(err)

	sources, err := serverapi.LoadStoreSources()
	essentials.Must(err)

	server := &serverapi.Server{
		AssetDir:   args.AssetDir,
		NumProxies: args.NumProxies,
		LocalMode:  args.LocalMode,
		DB:         dbInstance,
		Sources:    sources,
		StoreCache: serverapi.NewStoreCache(sources),
	}
	server.AddRoutes()
	mux := http.DefaultServeMux
	mux = serverapi.UncachedMux(mux)
	mux = serverapi.RateLimitMux(server, mux)
	http.ListenAndServe(args.Addr, mux)
}
