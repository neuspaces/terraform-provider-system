package acctest

import (
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
)

var current AccTest

// Current returns the current AccTest instance of the acceptance test run
func Current() AccTest {
	return current
}

// CurrentProviderConfigBlock returns the current configuration as a Terraform provider block
// Deprecated
func CurrentProviderConfigBlock() tfbuild.FileElement {
	return tfbuild.Provider(provider.Name)
	//return providerConfigBlock(current.Config)
}
