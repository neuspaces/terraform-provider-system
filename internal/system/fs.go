package system

import (
	"context"
	"io"
	"io/fs"
)

// ReadFS provides read access to a file system.
// ReadFS has support for context.Context in addition to the similar fs.FS.
type ReadFS interface {
	Open(ctx context.Context, name string) (fs.File, error)
}

// WriteFS provides write access to a file system.
type WriteFS interface {
	// Create creates or truncates the named file. If the file already exists,
	// it is truncated. If the file does not exist, it is created.
	// The size of the file must be known in advance.
	Create(ctx context.Context, fileInfo fs.FileInfo) (WriteFile, error)
}

type StatFS interface {
	Stat(ctx context.Context, name string) (fs.FileInfo, error)
}

// WriteFile provides write access to a single file
type WriteFile interface {
	io.Writer
	io.Closer
}
