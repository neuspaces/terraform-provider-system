package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"time"
)

const (
	SchemaAttrConnectionHost        = "host"
	SchemaAttrConnectionHostKey     = "host_key"
	SchemaAttrConnectionPort        = "port"
	SchemaAttrConnectionUser        = "user"
	SchemaAttrConnectionPassword    = "password"
	SchemaAttrConnectionPrivateKey  = "private_key"
	SchemaAttrConnectionCertificate = "certificate"

	SchemaAttrConnectionBastionHost        = "bastion_host"
	SchemaAttrConnectionBastionHostKey     = "bastion_host_key"
	SchemaAttrConnectionBastionPort        = "bastion_port"
	SchemaAttrConnectionBastionUser        = "bastion_user"
	SchemaAttrConnectionBastionPassword    = "bastion_password"
	SchemaAttrConnectionBastionPrivateKey  = "bastion_private_key"
	SchemaAttrConnectionBastionCertificate = "bastion_certificate"

	SchemaAttrConnectionTimeout       = "timeout"
	SchemaAttrConnectionAgent         = "agent"
	SchemaAttrConnectionAgentIdentity = "agent_identity"
)

func providerSchemaConnection(attrPath attrPath) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		SchemaAttrConnectionUser: {
			Description: "The user that should be used to connect to the remote ssh server. Defaults to `root`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionPassword: {
			Description: fmt.Sprintf("The password that should be used to authenticate with the remote ssh server. Mutually exclusive with `%[2]s`.", SchemaAttrConnectionPassword, SchemaAttrConnectionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionPrivateKey).String(),
			},
		},
		SchemaAttrConnectionPrivateKey: {
			Description: fmt.Sprintf("The SSH private key to authenticate with the remote ssh server. The key can be provided as text or loaded from a file using the `file` function. The key must not be encrypted. Mutually exclusive with `%[1]s`.", SchemaAttrConnectionPassword, SchemaAttrConnectionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionPassword).String(),
			},
		},
		SchemaAttrConnectionCertificate: {
			Description: fmt.Sprintf("The ssh user certificate to authenticate with the remote ssh server. The certificate can be provided as text or loaded from a file using the `file` function. Must be used with in conjunction with `%[2]s`. Mutually exclusive with `%[1]s`.", SchemaAttrConnectionPassword, SchemaAttrConnectionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionPassword).String(),
			},
		},
		SchemaAttrConnectionHost: {
			Description: "The host of the remote ssh server to connect to.",
			Type:        schema.TypeString,
			Required:    true,
		},
		SchemaAttrConnectionHostKey: {
			Description: "The public key or the CA certificate of the remote ssh host to verify the remote authenticity.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionPort: {
			Description: "The port of the remote ssh server to connect to. Defaults to `22`.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		SchemaAttrConnectionTimeout: {
			Description:      "The timeout to wait for the connection to be established. Should be provided as a string like `30s` or `5m`. Defaults to 5 minutes.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validate.Duration(),
		},
		SchemaAttrConnectionAgent: {
			Description: "Set to `false` to disable using an ssh agent to authenticate. Defaults to `true`.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		SchemaAttrConnectionAgentIdentity: {
			Description: "The preferred identity from the ssh agent for authentication.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionBastionHost: {
			Description: "Setting this attribute enables a bastion connection to the remote ssh server. The host of the bastion ssh server through which to connect to the remote ssh server.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionBastionHostKey: {
			Description: "The public key or the CA certificate of the bastion ssh host to verify the bastion host authenticity.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionBastionPort: {
			Description: "The port of the bastion ssh server to connect to. Defaults to `22`.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		SchemaAttrConnectionBastionUser: {
			Description: "The user that should be used to connect to the bastion ssh server. Defaults to `root`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		SchemaAttrConnectionBastionPassword: {
			Description: fmt.Sprintf("The password that should be used to authenticate. Mutually exclusive with `%[2]s`.", SchemaAttrConnectionBastionPassword, SchemaAttrConnectionBastionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionBastionPrivateKey).String(),
			},
		},
		SchemaAttrConnectionBastionPrivateKey: {
			Description: fmt.Sprintf("The SSH private key to authenticate with the bastion ssh server. The key can be provided as text or loaded from a file using the `file` function. The key must not be encrypted. Mutually exclusive with `%[1]s`.", SchemaAttrConnectionBastionPassword, SchemaAttrConnectionBastionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionBastionPassword).String(),
			},
		},
		SchemaAttrConnectionBastionCertificate: {
			Description: fmt.Sprintf("The ssh user certificate to authenticate with the bastion ssh server. The certificate can be provided as text or loaded from a file using the `file` function. Must be used with in conjunction with `%[2]s`. Mutually exclusive with `%[1]s`.", SchemaAttrConnectionBastionPassword, SchemaAttrConnectionBastionPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrConnectionBastionPassword).String(),
			},
		},
	}
}

// expandSchemaConnection expands attributes of the connection block into two SchemaSsh instances.
// The first SchemaSsh represents the remote ssh server and the second SchemaSsh represents the optional bastion ssh server.
func expandSchemaConnection(v interface{}) (*SchemaSsh, *SchemaSsh, error) {
	d, err := expandListSingle(v)
	if err != nil {
		return nil, nil, err
	}

	// Remote ssh server
	r := &SchemaSsh{
		User:            d[SchemaAttrConnectionUser].(string),
		Password:        d[SchemaAttrConnectionPassword].(string),
		PrivateKey:      d[SchemaAttrConnectionPrivateKey].(string),
		Certificate:     d[SchemaAttrConnectionCertificate].(string),
		Host:            d[SchemaAttrConnectionHost].(string),
		HostKey:         d[SchemaAttrConnectionHostKey].(string),
		Port:            d[SchemaAttrConnectionPort].(int),
		Agent:           d[SchemaAttrConnectionAgent].(bool),
		AgentIdentities: []string{},
	}

	if timeoutStr := d[SchemaAttrConnectionTimeout].(string); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, nil, err
		}
		r.Timeout = timeout
	}

	// Bastion ssh server
	var b *SchemaSsh
	if d[SchemaAttrConnectionBastionHost].(string) != "" {
		b = &SchemaSsh{
			User:        d[SchemaAttrConnectionBastionUser].(string),
			Password:    d[SchemaAttrConnectionBastionPassword].(string),
			PrivateKey:  d[SchemaAttrConnectionBastionPrivateKey].(string),
			Certificate: d[SchemaAttrConnectionBastionCertificate].(string),
			Host:        d[SchemaAttrConnectionBastionHost].(string),
			HostKey:     d[SchemaAttrConnectionBastionHostKey].(string),
			Port:        d[SchemaAttrConnectionBastionPort].(int),

			// Shared fields
			Agent:           r.Agent,
			AgentIdentities: r.AgentIdentities,
			Timeout:         r.Timeout,
		}
	}

	return r, b, nil
}
