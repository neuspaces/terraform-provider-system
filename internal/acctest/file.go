package acctest

import (
	"bytes"
	"io"
	"os"
)

// fileReadAll opens the named file for reading, reads its contents into a bytes.Buffer and returns an io.Reader on the buffer.
// The returned io.Reader may only be read once.
func fileReadAll(name string) (io.Reader, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
