package main

type Loader interface {
	Load() bool
	Filter() Filter
	Err() error
}

type Filter struct {
	Exception bool
	Domain    string
}
