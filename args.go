package main

import "flag"

type Args struct {
	AssetDir string
	DataDir  string
	Addr     string
}

func (a *Args) Add() {
	flag.StringVar(&a.AssetDir, "assets", "assets", "web asset directory")
	flag.StringVar(&a.DataDir, "data", "data", "data store directory")
	flag.StringVar(&a.Addr, "addr", ":8080", "address to listen on")
}
