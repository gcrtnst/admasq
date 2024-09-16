package main

import (
	"math/rand/v2"
	"os"
	"path/filepath"
)

type BlobCache struct {
	Root string
}

func (c *BlobCache) Create() (*os.File, string, error) {
	return c.create(rand.UintN)
}

func (c *BlobCache) create(rand func(n uint) uint) (*os.File, string, error) {
	const prefix = "data_"
	const randLen = 8
	const randTbl = "0123456789abcdefghijklmnopqrstuvwxyz"

	var nameBuf [len(prefix) + randLen]byte
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
