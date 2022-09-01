package limited

import "io"

// NewWriter returns an ioWriter that writes to w
// but stops with EOF after n bytes.
// The underlying implementation is a *Writer.
func NewWriter(w io.Writer, n int64) io.Writer { return &Writer{w, n} }

// A Writer writes to W but limits the amount of
// data returned to just N bytes. Each call to Write
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying W returns EOF.
type Writer struct {
	W io.Writer // underlying writer
	N int64     // max bytes remaining
}

func (l *Writer) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}
