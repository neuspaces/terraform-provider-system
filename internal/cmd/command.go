package cmd

import (
	"io"
)

type Command interface {
	Command() string

	// Stdin is an io.Reader for the standard input
	// Stdin is used by the executing System
	Stdin() io.Reader

	// Stdout is an io.Writer for the standard output
	// Stdout is used by the executing System
	Stdout() io.Writer

	// Stderr is an io.Writer for the standard error
	// Stderr is used by the executing System
	Stderr() io.Writer
}

type command struct {
	commandFunc func() string

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	result Result
}

var _ Command = &command{}

type CommandOption func(*command)

func Stdin(stdin io.Reader) CommandOption {
	return func(c *command) {
		c.stdin = stdin
	}
}

func Stdout(stdout io.Writer) CommandOption {
	return func(c *command) {
		c.stdout = stdout
	}
}

func Stderr(stderr io.Writer) CommandOption {
	return func(c *command) {
		c.stderr = stderr
	}
}

func Passthrough(parent Command) CommandOption {
	return func(c *command) {
		c.stdin = parent.Stdin()
		c.stdout = parent.Stdout()
		c.stderr = parent.Stderr()
	}
}

func CombinedOutput(out io.Writer) CommandOption {
	syncOut := newSyncWriter(out)

	return func(c *command) {
		c.stdout = syncOut
		c.stderr = syncOut
	}
}

func NewCommand(cmd string, opts ...CommandOption) Command {
	return NewCommandWithFunc(func() string { return cmd }, opts...)
}

func NewCommandWithFunc(cmdFunc func() string, opts ...CommandOption) Command {
	c := &command{
		commandFunc: cmdFunc,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *command) Command() string {
	return c.commandFunc()
}

func (c *command) Stdin() io.Reader {
	return c.stdin
}

func (c *command) Stdout() io.Writer {
	return c.stdout
}

func (c *command) Stderr() io.Writer {
	return c.stderr
}

func (c *command) Complete(result Result) {
	c.result = result
}

func (c *command) Result() Result {
	return c.result
}
