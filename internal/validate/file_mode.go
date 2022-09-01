package validate

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
)

var regexpFileMode = regexp.MustCompile(`^[0-7]{3}$`)

func FileMode() schema.SchemaValidateDiagFunc {
	return StringMatch(regexpFileMode, "invalid file mode")
}
