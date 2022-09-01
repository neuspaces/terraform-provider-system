package validate

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"net/url"
)

func ExpectUrl(val interface{}, path cty.Path) (*url.URL, diag.Diagnostics) {
	strVal, err := expectString(val, path)
	if err != nil {
		return nil, err
	}

	urlVal, urlErr := url.Parse(strVal)
	if urlErr != nil {
		return nil, []diag.Diagnostic{
			{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("expected a valid url, got %v", val),
				Detail:        urlErr.Error(),
				AttributePath: path,
			},
		}
	}

	return urlVal, nil
}
