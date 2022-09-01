package acctest

import (
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"testing"
)

type Targets []Target

type Target struct {
	ConfigTarget

	// Id which uniquely identifies the target
	Id string

	// BasePath is the path on the remote system in which all tests are executed
	BasePath string

	// Prefix is used for naming resource and contains the Id
	Prefix string

	Provider *provider.Provider
}

// Foreach is a convenience function to run a test for multiple targets
func (targets Targets) Foreach(t *testing.T, f func(*testing.T, Target)) {
	for _, target := range targets {
		t.Run(target.Id, func(t *testing.T) {
			f(t, target)
		})
	}
}
