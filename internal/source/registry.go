package source

import (
	"fmt"
	"net/url"
)

// Registry allows registering multiple Client by scheme
type Registry struct {
	clients       map[string]Client
	defaultScheme string
}

type RegistryOption func(*Registry) error

func NewRegistry(opts ...RegistryOption) (*Registry, error) {
	var err error

	r := &Registry{
		clients: map[string]Client{},
	}

	for _, opt := range opts {
		err = opt(r)
		if err != nil {
			return nil, err
		}
	}

	// Validate default schema is registered
	if r.defaultScheme != "" {
		if _, hasDefaultScheme := r.clients[r.defaultScheme]; !hasDefaultScheme {
			return nil, ErrRegistryOptions.WithCause(fmt.Errorf("no client registered for default scheme %s", r.defaultScheme))
		}
	}

	return r, nil
}

func WithClients(clients ...Client) RegistryOption {
	return func(r *Registry) error {
		for _, c := range clients {
			opt := WithClient(c)
			err := opt(r)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func WithClient(c Client) RegistryOption {
	return func(r *Registry) error {
		schemes := c.Schemes()
		for _, scheme := range schemes {
			if _, hasScheme := r.clients[scheme]; hasScheme {
				return ErrRegistryOptions.WithCause(fmt.Errorf("duplicate scheme not allowed: %s", scheme))
			}
			r.clients[scheme] = c
		}
		return nil
	}
}

func WithDefaultScheme(scheme string) RegistryOption {
	return func(r *Registry) error {
		if scheme == "" {
			return ErrRegistryOptions.WithCause(fmt.Errorf("default scheme must not be empty"))
		}
		r.defaultScheme = scheme
		return nil
	}
}

// Open returns a source for the provided raw url
func (r *Registry) Open(rawurl string) (Source, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	return r.OpenUrl(u)
}

// OpenUrl returns a source for the provided url
func (r *Registry) OpenUrl(u *url.URL) (Source, error) {
	// Copy url struct
	localU := *u

	// Fallback to default scheme if provided
	if localU.Scheme == "" && r.defaultScheme != "" {
		localU.Scheme = r.defaultScheme
	}

	if localU.Scheme == "" {
		return nil, ErrRegistryOpen.WithCause(fmt.Errorf("url does not contain a scheme"))
	}

	// Lookup source from schema
	client, clientExists := r.clients[localU.Scheme]
	if !clientExists {
		return nil, ErrRegistryOpen.WithCause(fmt.Errorf(`scheme "%s" is not supported`, localU.Scheme))
	}

	return client.Open(&localU)
}
