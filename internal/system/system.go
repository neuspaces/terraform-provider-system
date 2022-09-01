package system

import (
	"context"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"io"
)

// System is an interface to access a file system of and execute commands on a system.
type System interface {
	io.Closer
	FileSystem
	Executor
}

// FileSystem provides read and write access to a file system.
type FileSystem interface {
	ReadFS
	WriteFS
	StatFS
}

type Executor interface {
	Execute(context.Context, cmd.Command) (cmd.Result, error)
}
