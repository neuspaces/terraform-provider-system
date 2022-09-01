package system

import (
	"bytes"
	"context"
	cmd2 "github.com/neuspaces/terraform-provider-system/internal/cmd"
)

type ExecutionResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Execute runs a command on a system and returns an ExecutionResult.
// Execute implicitly records stdout and stderr and provides them in the ExecutionResult.
// Execute provides a higher abstraction than Command.
func Execute(ctx context.Context, system System, command string) (*ExecutionResult, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := cmd2.NewCommand(command, cmd2.Stdout(stdout), cmd2.Stderr(stderr))

	cmdResult, err := system.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}

	execResult := &ExecutionResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: cmdResult.ExitCode(),
	}

	return execResult, nil
}
