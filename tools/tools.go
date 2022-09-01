//go:build tools
// +build tools

package tools

import (
	// https://github.com/hashicorp/terraform-plugin-docs
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"

	// https://github.com/bflad/tfproviderlint
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
)
