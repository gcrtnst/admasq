package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSimpleParserParseEmpty(t *testing.T) {
	r := strings.NewReader("")
	p := NewSimpleParser(r)
	HelpSimpleParserTest(t, p, false, 0, "", nil)
}

func TestSimpleParserParseNormal(t *testing.T) {
	r := strings.NewReader("# begin\n1.example.com\n\n2.example.com\n# end")
	p := NewSimpleParser(r)
	HelpSimpleParserTest(t, p, true, 2, "1.example.com", nil)
	HelpSimpleParserTest(t, p, true, 4, "2.example.com", nil)
	HelpSimpleParserTest(t, p, false, 5, "", nil)
}

func TestSimpleParserParseCRLF(t *testing.T) {
	r := strings.NewReader("1.example.com\r\n2.example.com\r\n")
	p := NewSimpleParser(r)
	HelpSimpleParserTest(t, p, true, 1, "1.example.com", nil)
	HelpSimpleParserTest(t, p, true, 2, "2.example.com", nil)
	HelpSimpleParserTest(t, p, false, 2, "", nil)
}

func TestSimpleParserReadError(t *testing.T) {
	err := errors.New("test")
	r := &ErrorReader{Err: err}
	p := NewSimpleParser(r)
	HelpSimpleParserTest(t, p, false, 0, "", err)
}

func HelpSimpleParserTest(t *testing.T, p *SimpleParser, wantOK bool, wantLine int, wantDomain string, wantErr error) {
	t.Helper()

	gotOK := p.Parse()
	if gotOK != wantOK {
		t.Errorf("ok: expected %t, got %t", wantOK, gotOK)
	}

	gotLine := p.Line()
	if gotLine != wantLine {
		t.Errorf("p.Line(): expected %d, got %d", wantLine, gotLine)
	}

	gotDomain := p.Domain()
	if gotDomain != wantDomain {
		t.Errorf("p.Domain(): expected %q, got %q", wantDomain, gotDomain)
	}

	gotErr := p.Err()
	if gotErr != wantErr {
		t.Errorf("p.Err(): expected %#v, got %#v", wantErr, gotErr)
	}

	if t.Failed() {
		t.FailNow()
	}
}

func TestParseSimpleLine(t *testing.T) {
	tt := []struct {
		in   []byte
		want string
	}{
		{
			in:   nil,
			want: "",
		},
		{
			in:   []byte("   "),
			want: "",
		},
		{
			in:   []byte("# comment"),
			want: "",
		},
		{
			in:   []byte("example.com"),
			want: "example.com",
		},
		{
			in:   []byte("   example.com"),
			want: "example.com",
		},
		{
			in:   []byte("example.com   "),
			want: "example.com",
		},
		{
			in:   []byte("\texample.com\t"),
			want: "example.com",
		},
		{
			in:   []byte("example.com   # comment"),
			want: "example.com",
		},
		{
			in:   []byte("   s p a c e   "),
			want: "s p a c e",
		},
	}

	for _, tc := range tt {
		got := ParseSimpleLine(tc.in)
		if got != tc.want {
			t.Errorf("%q: expected %q, got %q", tc.in, tc.want, got)
		}
	}
}

type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (int, error) {
	return 0, r.Err
}
