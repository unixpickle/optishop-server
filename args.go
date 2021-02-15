package main

import "flag"

type Args struct {
	AssetDir   string
	DataDir    string
	Addr       string
	NumProxies int
	LocalMode  bool
}

func (a *Args) Add() {
	flag.StringVar(&a.AssetDir, "assets", "assets", "web asset directory")
	flag.StringVar(&a.DataDir, "data", "data", "data store directory")
	flag.StringVar(&a.Addr, "addr", ":8080", "address to listen on")
	flag.IntVar(&a.NumProxies, "proxies", 0, "number of reverse proxies before this endpoint, "+
		"for rate-limiting")
	flag.BoolVar(&a.LocalMode, "local", false, "provide a user-free front-end")
}
