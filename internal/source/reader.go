package source

import (
	"hash"
	"io"
)

type readerFunc func([]byte) (int, error)

var _ io.Reader = readerFunc(func(b []byte) (int, error) {
	return 0, nil
})

func (r readerFunc) Read(p []byte) (n int, err error) {
	return r(p)
}

type readCloser struct {
	reader io.Reader
	closer io.Closer
}

var _ io.ReadCloser = &readCloser{}

func (r *readCloser) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		return 0, io.EOF
	}

	return r.reader.Read(p)
}

func (r *readCloser) Close() error {
	if r.closer == nil {
		return nil
	}
	return r.closer.Close()
}

func newHashReader(h hash.Hash, rc io.ReadCloser) io.ReadCloser {
	tee := io.TeeReader(rc, h)
	return &readCloser{
		reader: tee,
		closer: rc,
	}
}

// hashFromReader calculates the hash.Hash for a byte stream provided by an io.Reader.
// hashFromReader reads from the io.Reader until io.EOF is reached.
func hashFromReader(h hash.Hash, r io.Reader) ([]byte, error) {
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
