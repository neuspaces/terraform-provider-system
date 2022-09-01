package local

import (
	"context"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io/fs"
	"os"
	"os/exec"
)

type System struct {
}

// System implements system.System
var _ system.System = &System{}

func NewSystem() system.System {
	return &System{}
}

func (s *System) Open(ctx context.Context, name string) (fs.File, error) {
	return os.Open(name)
}

func (s *System) Close() error {
	return nil
}

func (s *System) Create(ctx context.Context, fileInfo fs.FileInfo) (system.WriteFile, error) {
	return os.OpenFile(fileInfo.Name(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
}

func (s *System) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (s *System) Execute(ctx context.Context, c cmd.Command) (cmd.Result, error) {
	command := c.Command()
	execCmd := exec.CommandContext(ctx, "sh", "-c", command)

	execCmd.Stdin = c.Stdin()
	execCmd.Stdout = c.Stdout()
	execCmd.Stderr = c.Stderr()

	err := execCmd.Run()
	if err != nil {
		if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
			return cmd.NewResult(exitErr.ExitCode()), nil
		} else {
			return nil, err
		}
	}

	return cmd.NewResult(0), nil
}
