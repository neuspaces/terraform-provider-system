package provider

import (
	"encoding/hex"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/hashutil"
	"strings"
)

func dataIdFromAttrValues(a ...interface{}) (string, error) {
	id := fmt.Sprintf(strings.TrimRight(strings.Repeat("%s|", len(a)), "|"), a...)
	idSum, err := hashutil.Sha1Bytes([]byte(id))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(idSum), nil
}
