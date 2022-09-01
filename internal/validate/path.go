package validate

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func AbsolutePath() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		strVal, err := expectString(val, path)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(strVal, "/") {
			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("expected absolute path"),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}
