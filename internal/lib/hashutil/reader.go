package hashutil

import (
	"hash"
	"io"
)

// FromReader calculates the hash.Hash for a byte stream provided by an io.Reader
func FromReader(h hash.Hash, r io.Reader) ([]byte, error) {
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
