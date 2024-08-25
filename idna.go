package main

import (
	"strconv"

	"golang.org/x/net/idna"
)

func IDNAToASCII(s string) (string, error) {
	t, err := idna.Lookup.ToASCII(s)
	if err != nil {
		err = &IDNAError{
			Domain: s,
			Err:    err,
		}
	}
	return t, err
}

type IDNAError struct {
	Domain string
	Err    error
}

func (e *IDNAError) Error() string {
	b := []byte("domain ")
	b = strconv.AppendQuote(b, e.Domain)
	b = append(b, ": "...)
	b = append(b, e.Err.Error()...)
	return string(b)
}

func (e *IDNAError) Unwrap() error {
	return e.Err
}
