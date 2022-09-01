package etag

import (
	"fmt"
	"regexp"
)

const HeaderKey = "etag"

// Header represents an HTTP entity tag (ETag)
// https://datatracker.ietf.org/doc/html/rfc7232#section-2.3
type Header struct {
	ETag string

	// Weak is false if the ETag supports strong validation. A strongly validating ETag match indicates that the content of the two resource representations is byte-for-byte identical.
	Weak bool
}

func (h *Header) String() string {
	e := `"` + h.ETag + `"`

	if h.Weak {
		e = `W/` + e
	}

	return e
}

var regexpHeader = regexp.MustCompile(`^(W\/)?"(.*)"$`)

func Parse(header string) (*Header, error) {
	m := regexpHeader.FindStringSubmatch(header)
	if m == nil || len(m) != 3 {
		return nil, fmt.Errorf("invalid etag header: %s", header)
	}

	h := &Header{
		ETag: m[2],
		Weak: m[1] != "",
	}

	return h, nil
}
