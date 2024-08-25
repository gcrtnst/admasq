package main

import "strconv"

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
