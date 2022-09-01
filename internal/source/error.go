package source

import "strings"

type Error struct {
	msg   string
	cause error
}

var _ error = &Error{}

func (e *Error) Error() string {
	var parts []string

	if e.msg != "" {
		parts = append(parts, e.msg)
	}

	if e.cause != nil {
		parts = append(parts, e.cause.Error())
	}

	return strings.Join(parts, ": ")
}

func (e *Error) WithCause(cause error) *Error {
	return &Error{
		msg:   e.msg,
		cause: cause,
	}
}

// Error implements the Unwrap method
var _ interface{ Unwrap() error } = &Error{}

func (e *Error) Unwrap() error {
	return e.cause
}

// Error implements the interface{ Is(error) bool } used by pkg/errors.Is
var _ interface{ Is(error) bool } = &Error{}

// Is returns true, if target error is an equivalent error.
// For errors to be equivalent, msg must equal if not empty.
// Refer to https://golang.org/pkg/errors/#Is
func (e *Error) Is(target error) bool {
	if targetE, targetIsError := target.(*Error); targetIsError {
		return e.msg == targetE.msg
	}

	return false
}
