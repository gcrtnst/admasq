package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestBlobCacheOpen(t *testing.T) {
	c := &BlobCache{Root: t.TempDir()}

	f, name, err := c.Create()
	if err != nil {
		t.Fatal(err)
	}

	in := []byte("test")
	_, err = f.Write(in)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	f, err = c.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() // error ignored

	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, in) {
		t.Errorf("expected %q, got %q", in, got)
	}
}

func TestBlobCacheOpenInvalidName(t *testing.T) {
	c := &BlobCache{Root: t.TempDir()}
	f, err := c.Open("data-01234567")
	if f != nil {
		t.Errorf("f: expected nil, got %#v", f)
	}
	if err != ErrInvalidCachedBlobName {
		t.Errorf("err: expected %#v, got %#v", ErrInvalidCachedBlobName, err)
	}
}

func TestBlobCacheOpenNotFound(t *testing.T) {
	c := &BlobCache{Root: t.TempDir()}
	f, err := c.Open("data_01234567")
	if f != nil {
		t.Errorf("f: expected nil, got %#v", f)
	}
	if !os.IsNotExist(err) {
		t.Errorf("err: expected ErrNotExist, got %#v", err)
	}
}

func TestBlobCacheCreate(t *testing.T) {
	c := &BlobCache{Root: t.TempDir()}

	f, name, err := c.Create()
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	const pattern = `\Adata_[0-9a-z]{8}\z`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("name: expected to match %q, got %q", pattern, name)
	}
}

func TestCreateCacheBlobCreate(t *testing.T) {
	randList := []uint{0, 4, 8, 12, 16, 20, 24, 35}
	randIdx := 0
	randFunc := func(n uint) uint {
		const wantN = 36
		if n != wantN {
			t.Errorf("randFunc: n: expected %d, got %d", wantN, n)
		}

		randRet := randList[randIdx]
		randIdx++
		return randRet
	}

	root := t.TempDir()
	c := &BlobCache{Root: root}
	f, name, err := c.create(randFunc)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() // error ignored

	const wantName = "data_048cgkoz"
	if name != wantName {
		t.Errorf("name: expected %q, got %q", wantName, name)
	}

	wantFileName := filepath.Join(root, name)
	gotFileName := f.Name()
	if gotFileName != wantFileName {
		t.Errorf("f.Name(): expected %q, got %q", wantFileName, gotFileName)
	}

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	const wantStatSize = 0
	gotStatSize := stat.Size()
	if gotStatSize != wantStatSize {
		t.Errorf("stat.Size(): expected %d, got %d", wantStatSize, gotStatSize)
	}

	const wantStatMode = 0o644
	gotStatMode := stat.Mode()
	if gotStatMode != wantStatMode {
		t.Errorf("stat.Mode(): expected %O, got %O", wantStatMode, gotStatMode)
	}

	const wantStatIsDir = false
	gotStatIsDir := stat.IsDir()
	if gotStatIsDir != wantStatIsDir {
		t.Errorf("stat.IsDir(): expected %t, got %t", wantStatIsDir, gotStatIsDir)
	}

	_, err = f.Read(make([]byte, 4))
	if err == nil {
		t.Error("f.Read(): succeed unexpectedly")
	}

	_, err = f.Write([]byte("test"))
	if err != nil {
		t.Errorf("f.Write(): %s", err.Error())
	}
}

func TestCreateCacheBlobExists(t *testing.T) {
	rand := func(n uint) uint { return 0 }

	root := t.TempDir()
	path := filepath.Join(root, "data_00000000")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	c := &BlobCache{Root: root}
	f, name, err := c.create(rand)
	if f != nil {
		t.Error("f is not nil")
	}
	if name != "" {
		t.Error("name is not empty")
	}
	if !os.IsExist(err) {
		t.Error(err)
	}
}

func TestIsCachedBlobName(t *testing.T) {
	tt := []struct {
		in   string
		want bool
	}{
		{in: "data_01234567", want: true},   // ok
		{in: "data_01234569", want: true},   // ok '9'
		{in: "data_0123456a", want: true},   // ok 'a'
		{in: "data_0123456z", want: true},   // ok 'z'
		{in: "data_abcdefgh", want: true},   // ok lower
		{in: "", want: false},               // empty
		{in: "data_0123456", want: false},   // too short
		{in: "data_012345678", want: false}, // too long
		{in: "data-01234567", want: false},  // invalid prefix
		{in: "data_%1234567", want: false},  // invalid rand[0]
		{in: "data_0123%567", want: false},  // invalid rand[4]
		{in: "data_0123456%", want: false},  // invalid rand[7]
		{in: "data_0123456/", want: false},  // invalid rand (< '0')
		{in: "data_0123456:", want: false},  // invalid rand (> '9')
		{in: "data_0123456`", want: false},  // invalid rand (< 'a')
		{in: "data_0123456{", want: false},  // invalid rand (> 'z')
		{in: "data_ABCDEFGH", want: false},  // invalid rand (capital)
	}

	for _, tc := range tt {
		got := IsCachedBlobName(tc.in)
		if got != tc.want {
			t.Errorf("%q: expected %t, got %t", tc.in, tc.want, got)
		}
	}
}
