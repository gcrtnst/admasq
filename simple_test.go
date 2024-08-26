package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSimpleLoaderLoadEmpty(t *testing.T) {
	r := strings.NewReader("")
	l := NewSimpleLoader(r)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestSimpleLoaderLoadNormal(t *testing.T) {
	r := strings.NewReader("1.example.com\n2.example.com\n")
	l := NewSimpleLoader(r)
	HelpLoaderTest(t, l, true, Filter{Domain: "1.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "2.example.com"}, false)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestSimpleLoaderLoadException(t *testing.T) {
	r := strings.NewReader("1.example.com\n2.example.com\n")
	l := NewSimpleLoader(r)
	l.SetException(true)
	HelpLoaderTest(t, l, true, Filter{Exception: true, Domain: "1.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Exception: true, Domain: "2.example.com"}, false)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestSimpleLoaderIDNANormal(t *testing.T) {
	r := strings.NewReader("お名前.com\n")
	l := NewSimpleLoader(r)
	HelpLoaderTest(t, l, true, Filter{Domain: "xn--t8jx73hngb.com"}, false)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestSimpleLoaderIDNAError(t *testing.T) {
	r := strings.NewReader("--.com\n")
	l := NewSimpleLoader(r)

	HelpLoaderTest(t, l, true, Filter{Domain: "--.com"}, true)
	HelpResourceErrorTest(t, "l.Err()", l.Err(), "", 1)

	HelpLoaderTest(t, l, false, Filter{}, false)
}

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

	if p.Line != wantLine {
		t.Errorf("p.Line: expected %d, got %d", wantLine, p.Line)
	}

	if p.Domain != wantDomain {
		t.Errorf("p.Domain: expected %q, got %q", wantDomain, p.Domain)
	}

	if p.Err != wantErr {
		t.Errorf("p.Err: expected %#v, got %#v", wantErr, p.Err)
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
