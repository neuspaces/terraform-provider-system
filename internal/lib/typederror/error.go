package typederror

import (
	"errors"
	"fmt"
)

type ErrorType struct {
	parent *ErrorType
	label  string
}

// ErrorType implements the error interface
var _ error = &ErrorType{}

func (t *ErrorType) Error() string {
	if t.parent != nil {
		return fmt.Sprintf("%s: %s", t.parent.Error(), t.label)
	}
	return t.label
}

// ErrorType implements the Unwrap method
var _ interface{ Unwrap() error } = &ErrorType{}

func (t *ErrorType) Unwrap() error {
	if t == nil {
		return nil
	}
	return t.parent
}

// typedError implements the interface{ Is(error) bool } used by pkg/errors.Is
var _ interface{ Is(error) bool } = &ErrorType{}

// Is returns true, if and only if the label and parent of two ErrorType equal.
func (t *ErrorType) Is(target error) bool {
	if t == nil {
		return false
	} else if targetT, ok := target.(*ErrorType); ok {
		return t.label == targetT.label && t.parent == targetT.parent
	}
	return false
}

func New(label string, parent *ErrorType) *ErrorType {
	return &ErrorType{
		label:  label,
		parent: parent,
	}
}

func NewRoot(label string) *ErrorType {
	return New(label, nil)
}

func (t *ErrorType) Raise(err error) error {
	return &typedError{
		t:   t,
		err: err,
	}
}

type typedError struct {
	t   *ErrorType
	err error
}

// typedError implements the error interface
var _ error = &typedError{}

func (e *typedError) Error() string {
	if e.t != nil && e.err == nil {
		return e.t.Error()
	} else if e.t == nil && e.err != nil {
		return e.err.Error()
	}
	return fmt.Sprintf("%s: %s", e.t.Error(), e.err.Error())
}

// typedError implements the Unwrap method
var _ interface{ Unwrap() error } = &typedError{}

func (e *typedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// typedError implements the interface{ Is(error) bool } to be used by pkg/errors.Is
var _ interface{ Is(error) bool } = &typedError{}

// Is returns true, if the type field of the typedError instances are equal.
func (e *typedError) Is(target error) bool {
	if e == nil {
		return false
	} else if targetE, ok := target.(*typedError); ok {
		return e.t == targetE.t
	} else if targetT, ok := target.(*ErrorType); ok {
		return errors.Is(e.t, targetT)
	}
	return false
}
