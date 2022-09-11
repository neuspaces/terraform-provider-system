package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io"
)

type Command interface {
	Command() string
}

type InputCommand interface {
	Command
	Stdin() io.Reader
}

type command struct {
	command string
}

var _ Command = &command{}

func (c *command) Command() string {
	return c.command
}

func NewCommand(c string) Command {
	return &command{
		command: c,
	}
}

type inputCommand struct {
	command string
	stdin   io.Reader
}

var _ InputCommand = &inputCommand{}

func (c *inputCommand) Command() string {
	return c.command
}

func (c *inputCommand) Stdin() io.Reader {
	return c.stdin
}

func NewInputCommand(c string, in io.Reader) InputCommand {
	return &inputCommand{
		command: c,
		stdin:   in,
	}
}

type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

func (c *CommandResult) StdoutString() string {
	return string(c.Stdout)
}

func (c *CommandResult) StderrString() string {
	return string(c.Stderr)
}

// Error returns an error if the CommandResult has a non-negative exit code
func (c *CommandResult) Error() error {
	if c.ExitCode != 0 {
		return fmt.Errorf("command returned with exit code %d", c.ExitCode)
	}
	return nil
}

type ExecuteCommandOptions struct {
	stdoutFunc func(writer io.Writer) io.Writer
	stderrFunc func(writer io.Writer) io.Writer
}

type ExecuteCommandOption func(*ExecuteCommandOptions)

func WithStdout() ExecuteCommandOption {
	return WithStdoutFunc(func(w io.Writer) io.Writer {
		return w
	})
}

func WithStdoutFunc(f func(writer io.Writer) io.Writer) ExecuteCommandOption {
	return func(o *ExecuteCommandOptions) {
		o.stdoutFunc = f
	}
}

func WithStderr() ExecuteCommandOption {
	return WithStderrFunc(func(w io.Writer) io.Writer {
		return w
	})
}

func WithStderrFunc(f func(writer io.Writer) io.Writer) ExecuteCommandOption {
	return func(o *ExecuteCommandOptions) {
		o.stderrFunc = f
	}
}

func ExecuteCommand(ctx context.Context, s system.System, c Command) (*CommandResult, error) {
	return ExecuteCommandWithOptions(ctx, s, c, WithStdout(), WithStderr())
}

func ExecuteCommandWithOptions(ctx context.Context, s system.System, c Command, opts ...ExecuteCommandOption) (*CommandResult, error) {
	// Options
	o := &ExecuteCommandOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var cmdCommandOptions []cmd.CommandOption

	// Stdout
	stdout := &bytes.Buffer{}
	if o.stdoutFunc != nil {
		cmdCommandOptions = append(cmdCommandOptions, cmd.Stdout(o.stdoutFunc(stdout)))
	}

	// Stderr
	stderr := &bytes.Buffer{}
	if o.stderrFunc != nil {
		cmdCommandOptions = append(cmdCommandOptions, cmd.Stderr(o.stderrFunc(stderr)))
	}

	// Optional stdin if command is InputCommand
	if cmdIn, ok := c.(InputCommand); ok {
		stdin := cmdIn.Stdin()
		if stdin != nil {
			cmdCommandOptions = append(cmdCommandOptions, cmd.Stdin(stdin))
		}
	}

	cmdString := c.Command()
	cmdCommand := cmd.NewCommand(cmdString, cmdCommandOptions...)

	cmdResult, err := s.Execute(ctx, cmdCommand)
	if err != nil {
		return nil, err
	}

	commandResult := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: cmdResult.ExitCode(),
	}

	// Debug log
	tflog.Debug(ctx, "command executed", map[string]interface{}{
		"cmd":        cmdString,
		"exitcode":   commandResult.ExitCode,
		"stdout":     string(truncateBytes(commandResult.Stdout, 4096)),
		"stdout_len": len(commandResult.Stdout),
		"stderr":     string(truncateBytes(commandResult.Stderr, 4096)),
		"stderr_len": len(commandResult.Stderr),
	})

	return commandResult, nil
}
