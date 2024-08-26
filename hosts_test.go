package main

import (
	"net/netip"
	"slices"
	"testing"
)

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
