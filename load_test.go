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

func HelpLoaderTest(t *testing.T, l Loader, wantOK bool, wantF Filter, wantHasErr bool) {
	t.Helper()

	gotOK := l.Load()
	if gotOK != wantOK {
		t.Errorf("ok: expected %t, got %t", wantOK, gotOK)
	}

	gotF := l.Filter()
	if gotF != wantF {
		t.Errorf("l.Filter(): expected %#v, got %#v", wantF, gotF)
	}

	gotErr := l.Err()
	gotHasErr := gotErr != nil
	if gotHasErr != wantHasErr {
		t.Errorf("l.Err() != nil: expected %t, got %t", wantHasErr, gotHasErr)
	}
}

func HelpResourceErrorTest(t *testing.T, name string, gotErr error, wantName string, wantLine int) {
	t.Helper()

	gotResErr, ok := gotErr.(*ResourceError)
	if !ok {
		t.Errorf("%s.(type): expected *ResourceError, got %T", name, gotErr)
	}

	if gotResErr.Name != wantName {
		t.Errorf("%s.Name: expected %q, got %q", name, wantName, gotResErr.Name)
	}

	if gotResErr.Line != wantLine {
		t.Errorf("%s.Line: expected %d, got %d", name, wantLine, gotResErr.Line)
	}

	if gotResErr.Err == nil {
		t.Errorf("%s.Err: expected non-nil error, got nil", name)
	}
}

type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (int, error) {
	return 0, r.Err
}
