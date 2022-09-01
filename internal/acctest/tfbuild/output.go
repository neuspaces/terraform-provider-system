package tfbuild

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// FileString returns the hcl.File as string
func FileString(file *hclwrite.File) string {
	if file == nil {
		return ""
	}
	s := string(file.Bytes())
	return s
}
