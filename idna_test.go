package main

import (
	"errors"
	"testing"
)

func TestIDNAToASCIINormal(t *testing.T) {
	gotS, gotErr := IDNAToASCII("お名前.com")

	const wantS = "xn--t8jx73hngb.com"
	if gotS != wantS {
		t.Errorf("s: expected %q, got %q", wantS, gotS)
	}

	if gotErr != nil {
		t.Errorf("err: expected nil, got %#v", gotErr)
	}
}

func TestIDNAToASCIIError(t *testing.T) {
	const inS = "--.com"
	gotS, gotErr := IDNAToASCII(inS)

	if gotS != inS {
		t.Errorf("s: expected %q, got %q", inS, gotS)
	}

	if gotIDNAError, ok := gotErr.(*IDNAError); !ok {
		t.Errorf("err.(type): expected *IDNAError, got %T", gotErr)
	} else {
		if gotIDNAError.Domain != inS {
			t.Errorf("err.Domain: expected %q, got %q", inS, gotIDNAError.Domain)
		}
		if gotIDNAError.Err == nil {
			t.Error("err.Err: expected non-nil error, got nil")
		}
	}
}

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
