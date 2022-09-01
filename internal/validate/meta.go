package validate

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// All returns a SchemaValidateDiagFunc which tests if the provided value
// passes all provided SchemaValidateDiagFunc
func All(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var allDiags diag.Diagnostics
		for _, validator := range validators {
			validatorDiags := validator(val, path)
			allDiags = append(allDiags, validatorDiags...)
		}
		return allDiags
	}
}

// Any returns a SchemaValidateDiagFunc which tests if the provided value
// passes at least one (any) of the provided SchemaValidateDiagFunc
func Any(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var allDiags diag.Diagnostics
		for _, validator := range validators {
			validatorDiags := validator(val, path)
			if len(validatorDiags) == 0 {
				return nil
			}
			allDiags = append(allDiags, validatorDiags...)
		}
		return allDiags
	}
}
