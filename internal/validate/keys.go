package validate

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/sshclient"
	"golang.org/x/crypto/ssh"
)

// PrivateKey validates if the value can be parsed using ssh.ParseRawPrivateKey.
// Supported private keys are unencrypted pem encoded RSA (PKCS#1), PKCS#8, DSA (OpenSSL), and ECDSA private keys.
func PrivateKey() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		strVal, diagErr := expectString(val, path)
		if diagErr != nil {
			return diagErr
		}

		_, err := ssh.ParsePrivateKey([]byte(strVal))
		if err != nil {
			if err.Error() == (&ssh.PassphraseMissingError{}).Error() {
				return []diag.Diagnostic{
					{
						Severity:      diag.Error,
						Summary:       fmt.Sprintf("password protected private key is not supported"),
						AttributePath: path,
					},
				}
			} else {
				return []diag.Diagnostic{
					{
						Severity:      diag.Error,
						Summary:       fmt.Sprintf("invalid private key format"),
						Detail:        err.Error(),
						AttributePath: path,
					},
				}
			}
		}

		return nil
	}
}

// Base64PublicKey validates if the value can be parsed using ssh.ParsePublicKey.
func Base64PublicKey() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		strVal, diagErr := expectString(val, path)
		if diagErr != nil {
			return diagErr
		}

		_, err := sshclient.ParseBase64PublicKey(strVal)

		if err != nil {
			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("invalid public key format"),
					Detail:        err.Error(),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}

// AuthorizedKey validates if the value can be parsed using ssh.ParsePublicKey.
func AuthorizedKey() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		strVal, diagErr := expectString(val, path)
		if diagErr != nil {
			return diagErr
		}

		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strVal))

		if err != nil {
			return []diag.Diagnostic{
				{
					Severity:      diag.Error,
					Summary:       fmt.Sprintf("invalid authorized key format"),
					Detail:        err.Error(),
					AttributePath: path,
				},
			}
		}

		return nil
	}
}
