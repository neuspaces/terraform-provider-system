package source

import (
	"io"
)

type Source interface {
	io.ReadCloser

	// Meta returns key-value pairs which are unique to the content returned by the source
	// The consumer of a source can infer a change of the content when Meta change
	Meta() (Meta, error)
}
