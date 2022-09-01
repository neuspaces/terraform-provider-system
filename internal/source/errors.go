package source

var (
	ErrRegistryOptions = &Error{msg: "invalid registry option"}

	ErrRegistryOpen = &Error{msg: "failed to open a source"}
)
