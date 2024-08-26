package main

import (
	"bytes"
	"errors"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestHostsLoaderLoadEmpty(t *testing.T) {
	r := strings.NewReader("")
	l := NewHostsLoader(r)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderLoadNormal(t *testing.T) {
	s := "" +
		"127.0.0.1 1.example.com\n" +
		"127.0.0.1 2.example.com 3.example.com 4.example.com\n" +
		"127.0.0.1 5.example.com 6.example.com\n" +
		"0.0.0.0   7.example.com\n"
	r := strings.NewReader(s)
	l := NewHostsLoader(r)
	HelpLoaderTest(t, l, true, Filter{Domain: "1.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "2.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "3.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "4.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "5.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "6.example.com"}, false)
	HelpLoaderTest(t, l, true, Filter{Domain: "7.example.com"}, false)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderIPError(t *testing.T) {
	r := strings.NewReader("x.x.x.x example.com")
	l := NewHostsLoader(r)

	HelpLoaderTest(t, l, true, Filter{}, true)
	gotErr := l.Err()
	if gotResErr, ok := gotErr.(*ResourceError); !ok {
		t.Errorf("l.Err().(type): expected *ResourceError, got %T", gotErr)
	} else {
		const wantResErrName = ""
		if gotResErr.Name != wantResErrName {
			t.Errorf("l.Err().Name: expected %q, got %q", wantResErrName, gotResErr.Name)
		}

		const wantResErrLine = 1
		if gotResErr.Line != wantResErrLine {
			t.Errorf("l.Err().Line: expected %d, got %d", wantResErrLine, gotResErr.Line)
		}

		if gotResErr.Err == nil {
			t.Error("l.Err().Err: expected non-nil error, got nil")
		}
	}

	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderRedirectError(t *testing.T) {
	r := strings.NewReader("192.168.0.1 example.com")
	l := NewHostsLoader(r)

	HelpLoaderTest(t, l, true, Filter{}, true)

	wantErr := &ResourceError{
		Line: 1,
		Err: &HostsIPError{
			IP: netip.AddrFrom4([4]byte{192, 168, 0, 1}),
		},
	}
	gotErr := l.Err()
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("l.Err(): expected %#v, got %#v", wantErr, gotErr)
	}

	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderNoHostError(t *testing.T) {
	r := strings.NewReader("127.0.0.1")
	l := NewHostsLoader(r)

	HelpLoaderTest(t, l, true, Filter{}, true)

	wantErr := &ResourceError{Line: 1, Err: ErrMissingHostname}
	gotErr := l.Err()
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("l.Err(): expected %#v, got %#v", wantErr, gotErr)
	}

	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderIDNANormal(t *testing.T) {
	r := strings.NewReader("127.0.0.1 お名前.com")
	l := NewHostsLoader(r)
	HelpLoaderTest(t, l, true, Filter{Domain: "xn--t8jx73hngb.com"}, false)
	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderIDNAError(t *testing.T) {
	r := strings.NewReader("127.0.0.1 --.com")
	l := NewHostsLoader(r)

	HelpLoaderTest(t, l, true, Filter{Domain: "--.com"}, true)
	HelpResourceErrorTest(t, "l.Err()", l.Err(), "", 1)

	HelpLoaderTest(t, l, false, Filter{}, false)
}

func TestHostsLoaderReadError(t *testing.T) {
	mockErr := errors.New("test")
	r := &ErrorReader{Err: mockErr}
	l := NewHostsLoader(r)

	HelpLoaderTest(t, l, false, Filter{}, true)
	gotErr := l.Err()
	if gotErr != mockErr {
		t.Errorf("l.Err(): expected %#v, got %#v", mockErr, gotErr)
	}
}

func TestHostsParserParseSingle(t *testing.T) {
	tt := []struct {
		name      string
		inBuf     []byte
		wantOK    bool
		wantLine  int
		wantIP    netip.Addr
		wantHosts []string
		wantIsErr bool
	}{
		{
			name:      "Empty",
			inBuf:     []byte{},
			wantOK:    false,
			wantLine:  0,
			wantIP:    netip.Addr{},
			wantHosts: nil,
			wantIsErr: false,
		},
		{
			name:      "Normal",
			inBuf:     []byte("127.0.0.1 example.com\n"),
			wantOK:    true,
			wantLine:  1,
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHosts: []string{"example.com"},
			wantIsErr: false,
		},
		{
			name:      "Warning",
			inBuf:     []byte("example.com\n"),
			wantOK:    true,
			wantLine:  1,
			wantIP:    netip.Addr{},
			wantHosts: nil,
			wantIsErr: true,
		},
		{
			name:      "Comment",
			inBuf:     []byte("# comment\n"),
			wantOK:    false,
			wantLine:  1,
			wantIP:    netip.Addr{},
			wantHosts: nil,
			wantIsErr: false,
		},
		{
			name:      "CRLF",
			inBuf:     []byte("127.0.0.1 example.com\r\n"),
			wantOK:    true,
			wantLine:  1,
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHosts: []string{"example.com"},
			wantIsErr: false,
		},
	}

	for _, tc := range tt {
		r := bytes.NewReader(tc.inBuf)
		p := NewHostsParser(r)
		gotOK := p.Parse()

		if gotOK != tc.wantOK {
			t.Errorf("%s: ok: expected %t, got %t", tc.name, tc.wantOK, gotOK)
		}

		if p.Line != tc.wantLine {
			t.Errorf("%s: p.Line: expected %d, got %d", tc.name, tc.wantLine, p.Line)
		}

		if !slices.Equal(p.Hosts, tc.wantHosts) {
			t.Errorf("%s: p.Hosts: expected %v, got %v", tc.name, tc.wantHosts, p.Hosts)
		}

		gotIsErr := p.Err != nil
		if gotIsErr != tc.wantIsErr {
			wantErr := "no error"
			if tc.wantIsErr {
				wantErr = "any error"
			}

			t.Errorf("%s: p.Err: expected %s, got %#v", tc.name, wantErr, p.Err)
		}
	}
}

func TestHostsParserParseMultiple(t *testing.T) {
	s := "127.0.0.1 example.com\n# comment\n0.0.0.0 another.example.com\n"
	r := strings.NewReader(s)
	p := NewHostsParser(r)
	var ok bool

	ok = p.Parse()
	if !ok {
		t.Fatal("first parse failed")
	}
	if p.IP != netip.AddrFrom4([4]byte{127, 0, 0, 1}) {
		t.Errorf("first parse: p.IP: expected 127.0.0.1, got %s", p.IP)
	}
	if p.Line != 1 {
		t.Errorf("first parse: p.Line: expected 1, got %d", p.Line)
	}
	if !slices.Equal(p.Hosts, []string{"example.com"}) {
		t.Errorf("first parse: p.Hosts: expected [example.com], got %v", p.Hosts)
	}
	if p.Err != nil {
		t.Errorf("first parse: p.Err: expected nil, got %#v", p.Err)
	}

	ok = p.Parse()
	if !ok {
		t.Fatal("second parse failed")
	}
	if p.Line != 3 {
		t.Errorf("first parse: p.Line: expected 3, got %d", p.Line)
	}
	if p.IP != netip.AddrFrom4([4]byte{0, 0, 0, 0}) {
		t.Errorf("second parse: p.IP: expected 0.0.0.0, got %s", p.IP)
	}
	if !slices.Equal(p.Hosts, []string{"another.example.com"}) {
		t.Errorf("second parse: p.Hosts: expected [another.example.com], got %v", p.Hosts)
	}
	if p.Err != nil {
		t.Errorf("second parse: p.Err: expected nil, got %#v", p.Err)
	}

	ok = p.Parse()
	if ok {
		t.Fatal("third parse unexpectedly success")
	}
	if p.Err != nil {
		t.Errorf("third parse: p.Err: expected nil, got %#v", p.Err)
	}
}

func TestHostsParserParseError(t *testing.T) {
	s := "\n\nexample.com"
	r := strings.NewReader(s)
	p := NewHostsParser(r)

	if !p.Parse() {
		t.Fatal("parse failed")
	}

	err, ok := p.Err.(*ResourceError)
	if !ok {
		t.Fatalf("err: expected *ResourceError, got %T", p.Err)
	}
	if err.Name != "" {
		t.Errorf("err.Name: expected \"\", got %q", err.Name)
	}
	if err.Line != 3 {
		t.Errorf("err.Line: expected 3, got %d", err.Line)
	}
	if err.Err == nil {
		t.Error("err.Err: expected any error, got nil")
	}
}

func TestHostsParserReadError(t *testing.T) {
	err := errors.New("test error")
	r := &ErrorReader{Err: err}
	p := NewHostsParser(r)

	ok := p.Parse()
	if ok {
		t.Fatal("parse unexpectedly success")
	}
	if p.Err != err {
		t.Errorf("p.Err: expected %#v, got %#v", err, p.Err)
	}
}

func TestParseHostsLine(t *testing.T) {
	tt := []struct {
		name      string
		inLine    []byte
		wantIP    netip.Addr
		wantHs    []string
		wantIsErr bool
	}{
		{
			name:      "Empty",
			inLine:    nil,
			wantIP:    netip.Addr{},
			wantHs:    nil,
			wantIsErr: false,
		},
		{
			name:      "IPOnly",
			inLine:    []byte("127.0.0.1"),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    nil,
			wantIsErr: false,
		},
		{
			name:      "IPHost",
			inLine:    []byte("127.0.0.1 example.com"),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    []string{"example.com"},
			wantIsErr: false,
		},
		{
			name:      "IPHostAlias",
			inLine:    []byte("127.0.0.1 example.com alias.example.com"),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    []string{"example.com", "alias.example.com"},
			wantIsErr: false,
		},
		{
			name:      "IPInvalid",
			inLine:    []byte("example.com"),
			wantIP:    netip.Addr{},
			wantHs:    nil,
			wantIsErr: true,
		},
		{
			name:      "WhiteSpaceTab",
			inLine:    []byte("127.0.0.1\texample.com"),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    []string{"example.com"},
			wantIsErr: false,
		},
		{
			name:      "WhiteSpaceExtra",
			inLine:    []byte("  127.0.0.1   example.com   "),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    []string{"example.com"},
			wantIsErr: false,
		},
		{
			name:      "Comment",
			inLine:    []byte("127.0.0.1 example.com # comment"),
			wantIP:    netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			wantHs:    []string{"example.com"},
			wantIsErr: false,
		},
	}

	for _, tc := range tt {
		gotIP, gotHs, gotErr := ParseHostsLine(tc.inLine)

		if gotIP != tc.wantIP {
			t.Errorf("%s: ip: expected %s, got %s", tc.name, tc.wantIP, gotIP)
		}

		if !slices.Equal(gotHs, tc.wantHs) {
			t.Errorf("%s: hs: expected %v, got %v", tc.name, tc.wantHs, gotHs)
		}

		gotIsErr := gotErr != nil
		if gotIsErr != tc.wantIsErr {
			wantErr := "no error"
			if tc.wantIsErr {
				wantErr = "any error"
			}

			t.Errorf("%s: err: expected %s, got %#v", tc.name, wantErr, gotErr)
		}
	}
}

func TestHostsIPErrorError(t *testing.T) {
	err := &HostsIPError{IP: netip.AddrFrom4([4]byte{192, 168, 0, 1})}
	got := err.Error()
	const want = "192.168.0.1 is neither a loopback address nor an unspecified address"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
