package main

import "flag"

type Args struct {
	StoreID  string
	AssetDir string
}

func (a *Args) Add() {
	flag.StringVar(&a.StoreID, "store", "", "store ID")
	flag.StringVar(&a.AssetDir, "assets", "assets", "web asset directory")
}
