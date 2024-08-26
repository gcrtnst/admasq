package main

import (
	"bytes"
	"errors"
	"net/netip"
	"slices"
	"strings"
	"testing"
)

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
