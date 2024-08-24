package main

import "testing"

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
