package cmd

import (
	"github.com/alessio/shellescape"
)

// Middleware wraps to a command similar to a http middleware.
type Middleware func(Command) Command

// ShMiddleware returns a Middleware which executes a command using the shell /bin/sh
func ShMiddleware() Middleware {
	return func(c Command) Command {
		return NewCommandWithFunc(func() string {
			return `/bin/sh -c ` + shellescape.Quote(c.Command())
		}, Passthrough(c))
	}
}

// SudoShMiddleware returns a Middleware which executes a command using `sudo`
func SudoShMiddleware() Middleware {
	return func(c Command) Command {
		return NewCommandWithFunc(func() string {
			return `sudo /bin/sh -c ` + shellescape.Quote(c.Command())
		}, Passthrough(c))
	}
}
