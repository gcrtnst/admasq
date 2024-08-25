package main

import (
	"errors"
	"testing"
)

func TestIDNAErrorError(t *testing.T) {
	err := &IDNAError{
		Domain: "example.com",
		Err:    errors.New("error message"),
	}

	const want = `domain "example.com": error message`
	got := err.Error()
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
