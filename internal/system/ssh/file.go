package ssh

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io"
	"io/fs"
)

// readFile implements fs.File for a known fs.FileInfo and an io.ReadCloser
type readFile struct {
	fileInfo   fs.FileInfo
	readCloser io.ReadCloser
}

var _ fs.File = &readFile{}

func (f *readFile) Stat() (fs.FileInfo, error) {
	return f.fileInfo, nil
}

func (f *readFile) Read(b []byte) (int, error) {
	return f.readCloser.Read(b)
}

func (f *readFile) Close() error {
	return f.readCloser.Close()
}

// newCatFileReader returns an io.ReadCloser which reads a file from a System using the `cat` command
func newCatFileReader(ctx context.Context, s system.System, name string) io.ReadCloser {
	// Create pipe: pipe reader is returned to the caller; pipe writer captures stdout
	pipeReader, pipeWriter := io.Pipe()
	catCmd := cmd.NewCommand(fmt.Sprintf(`cat '%s'`, name), cmd.Stdout(pipeWriter))

	go func() {
		res, err := s.Execute(ctx, catCmd)
		if err != nil {
			// Error will be returned by Read of the PipeReader
			_ = pipeWriter.CloseWithError(err)
			return
		}

		if rc := res.ExitCode(); rc != 0 {
			// Error will be returned by Read of the PipeReader
			_ = pipeWriter.CloseWithError(fmt.Errorf("non-zero exit code: %d", rc))
			return
		}

		// Read completed
		_ = pipeWriter.Close()
	}()

	// TODO implement a lazy reader; i.e. start reading on first read call; cat reader
	// TODO close if context is cancelled in separate goroutine; actually this should be the responsibility of Execute to close the Stdout

	return pipeReader
}
