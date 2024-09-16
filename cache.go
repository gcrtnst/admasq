package main

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidCachedBlobName = errors.New("invalid data file name in cache")

type BlobCache struct {
	Root string
}

func (c *BlobCache) Open(name string) (*os.File, error) {
	if !IsCachedBlobName(name) {
		return nil, ErrInvalidCachedBlobName
	}

	path := filepath.Join(c.Root, name)
	return os.Open(path)
}

func (c *BlobCache) Create() (*os.File, string, error) {
	return c.create(rand.UintN)
}

func (c *BlobCache) create(rand func(n uint) uint) (*os.File, string, error) {
	const prefix = "data_"
	const randLen = 8
	const randTbl = "0123456789abcdefghijklmnopqrstuvwxyz"
	const nameLen = len(prefix) + randLen

	var nameBuf [nameLen]byte
	copy(nameBuf[:], prefix)

	root := c.Root

	var err error
	for range 10000 {
		for i := len(prefix); i < len(nameBuf); i++ {
			nameBuf[i] = randTbl[rand(uint(len(randTbl)))]
		}

		var f *os.File
		name := string(nameBuf[:])
		path := filepath.Join(root, name)
		f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if !os.IsExist(err) {
			return f, name, err
		}
	}
	return nil, "", err
}

func IsCachedBlobName(name string) bool {
	const prefix = "data_"
	const randLen = 8
	const nameLen = len(prefix) + randLen

	if len(name) != nameLen || !strings.HasPrefix(name, prefix) {
		return false
	}

	for i := len(prefix); i < len(name); i++ {
		b := name[i]
		if (b < '0' || '9' < b) && (b < 'a' || 'z' < b) {
			return false
		}
	}

	return true
}
