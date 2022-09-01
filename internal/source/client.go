package source

import "net/url"

// Client opens a Source from a specific scheme
type Client interface {
	// Open returns a source for the provided URL
	Open(url *url.URL) (Source, error)

	// Schemes returns a list of schemes supported by the Client
	Schemes() []string
}
