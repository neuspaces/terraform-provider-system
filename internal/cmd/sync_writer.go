package cmd

import (
	"io"
	"sync"
)

type syncWriter struct {
	w  io.Writer
	mu sync.Mutex
}

var _ io.Writer = &syncWriter{}

func newSyncWriter(writer io.Writer) *syncWriter {
	return &syncWriter{
		w: writer,
	}
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}
