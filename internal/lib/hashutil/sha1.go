package hashutil

import (
	"bytes"
	"crypto/sha1"
)

func Sha1Bytes(d []byte) ([]byte, error) {
	return FromReader(sha1.New(), bytes.NewReader(d))
}
