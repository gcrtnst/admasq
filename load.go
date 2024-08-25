package main

import (
	"strconv"
)

type Loader interface {
	Load() bool
	Filter() Filter
	Err() error
}

type Filter struct {
	Exception bool
	Domain    string
}

type ResourceError struct {
	Name string
	Line int
	Err  error
}

func (e *ResourceError) Error() string {
	b := []byte(e.Name)

	if e.Line > 0 {
		if len(b) > 0 {
			b = append(b, ':')
		}
		b = strconv.AppendInt(b, int64(e.Line), 10)
	}

	w := e.Err.Error()
	if len(w) > 0 {
		if len(b) > 0 {
			b = append(b, ": "...)
		}
		b = append(b, w...)
	}

	return string(b)
}

func (e *ResourceError) Unwrap() error {
	return e.Err
}
