package validate

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
)

func StringMatch(r *regexp.Regexp, message string) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		strVal, err := expectString(val, path)
		if err != nil {
			return err
		}

		if ok := r.MatchString(strVal); !ok {
			var summary string

			if message != "" {
				summary = message
			} else {
				summary = "invalid value"
			}

			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       summary,
					Detail:        fmt.Sprintf("expected value to match regular expression %q", r),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}

func expectString(val interface{}, path cty.Path) (string, diag.Diagnostics) {
	strVal, isStr := val.(string)
	if !isStr {
		return "", []diag.Diagnostic{
			{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("expected type string"),
				AttributePath: path,
			},
		}
	}

	return strVal, nil
}
