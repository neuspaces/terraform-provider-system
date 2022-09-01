package validate

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

// Duration returns a schema.SchemaValidateDiagFunc which tests if the provided value
// is of type string and can be parsed as time.Duration
func Duration() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		_, err := expectDuration(val, path)
		if err != nil {
			return err
		}

		return nil
	}
}

// DurationAtLeast returns a schema.SchemaValidateDiagFunc if the provided value
// is of type string, can be parsed as time.Duration, and is at least min (inclusive)
func DurationAtLeast(min time.Duration) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		v, err := expectDuration(val, path)
		if err != nil {
			return err
		}

		if v < min {
			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("expected value to be at least %s, got %s", min.String(), v.String()),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}

// DurationAtMost returns a schema.SchemaValidateDiagFunc if the provided value
// is of type string, can be parsed as time.Duration, and is at most max (inclusive)
func DurationAtMost(max time.Duration) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		v, err := expectDuration(val, path)
		if err != nil {
			return err
		}

		if v > max {
			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("expected value to be at most %s, got %s", max.String(), v.String()),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}

func expectDuration(val interface{}, path cty.Path) (time.Duration, diag.Diagnostics) {
	strVal, err := expectString(val, path)
	if err != nil {
		return 0, err
	}

	durationVal, durationErr := time.ParseDuration(strVal)
	if durationErr != nil {
		return 0, []diag.Diagnostic{
			{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("invalid duration format: %s", strVal),
				Detail:        durationErr.Error(),
				AttributePath: path,
			},
		}
	}

	return durationVal, nil
}
