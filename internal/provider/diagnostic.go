package provider

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	internalDiagnosticDetail = "this is an internal provider error and should be reported as an issue via https://github.com/neuspaces/terraform-provider-system/issues"
)

func newDiagnostic(severity diag.Severity, summary string, detail string, path cty.Path) diag.Diagnostic {
	return diag.Diagnostic{
		Severity:      severity,
		Summary:       summary,
		Detail:        detail,
		AttributePath: path,
	}
}

func newShortDiagnostic(severity diag.Severity, summary string) diag.Diagnostics {
	return diag.Diagnostics{
		newDiagnostic(severity, summary, "", nil),
	}
}

func newDetailedDiagnostic(severity diag.Severity, summary string, detail string, path cty.Path) diag.Diagnostics {
	return diag.Diagnostics{
		newDiagnostic(severity, summary, detail, path),
	}
}

func newInternalErrorDiagnostic(summary string) diag.Diagnostics {
	return diag.Diagnostics{
		newDiagnostic(diag.Error, summary, internalDiagnosticDetail, nil),
	}
}

func newInternalUnexpectedTypeDiagnostic(expected string, actual interface{}) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("expected type %s, got unexpected type %T", expected, actual),
		Detail:   internalDiagnosticDetail,
	}}
}
