package main

import (
	"errors"
	"testing"
)

func TestResourceErrorError(t *testing.T) {
	tt := []struct {
		name string
		in   *ResourceError
		want string
	}{
		{
			name: "Empty",
			in: &ResourceError{
				Name: "",
				Line: 0,
				Err:  errors.New(""),
			},
			want: "",
		},
		{
			name: "Name",
			in: &ResourceError{
				Name: "hosts.txt",
				Line: 0,
				Err:  errors.New(""),
			},
			want: "hosts.txt",
		},
		{
			name: "Line",
			in: &ResourceError{
				Name: "",
				Line: 52149,
				Err:  errors.New(""),
			},
			want: "52149",
		},
		{
			name: "Err",
			in: &ResourceError{
				Name: "",
				Line: 0,
				Err:  errors.New("some error"),
			},
			want: "some error",
		},
		{
			name: "All",
			in: &ResourceError{
				Name: "hosts.txt",
				Line: 52149,
				Err:  errors.New("some error"),
			},
			want: "hosts.txt:52149: some error",
		},
	}

	for _, tc := range tt {
		got := tc.in.Error()
		if got != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, got)
		}
	}
}

type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (int, error) {
	return 0, r.Err
}
