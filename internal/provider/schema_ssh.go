package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"time"
)

const (
	SchemaAttrSshHost            = "host"
	SchemaAttrSshHostKey         = "host_key"
	SchemaAttrSshPort            = "port"
	SchemaAttrSshUser            = "user"
	SchemaAttrSshPassword        = "password"
	SchemaAttrSshPrivateKey      = "private_key"
	SchemaAttrSshCertificate     = "certificate"
	SchemaAttrSshTimeout         = "timeout"
	SchemaAttrSshAgent           = "agent"
	SchemaAttrSshAgentIdentity   = "agent_identity"
	SchemaAttrSshAgentIdentities = "agent_identities"
)

type SchemaSsh struct {
	User            string
	Password        string
	PrivateKey      string
	Certificate     string
	Host            string
	HostKey         string
	Port            int
	Timeout         time.Duration
	Agent           bool
	AgentIdentities []string
}

func providerSchemaSsh(attrPath attrPath, envPrefix string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		SchemaAttrSshUser: {
			Description: "The user that should be used to connect to the remote ssh server.",
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSshUser, envPrefix, nil),
		},
		SchemaAttrSshPassword: {
			Description: fmt.Sprintf("The password that should be used to authenticate with the remote ssh server. Mutually exclusive with `%[2]s`.", SchemaAttrSshPassword, SchemaAttrSshPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrSshPrivateKey).String(),
			},
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSshPassword, envPrefix, nil),
		},
		SchemaAttrSshPrivateKey: {
			Description: fmt.Sprintf("The SSH private key to authenticate with the remote ssh server. The key can be provided as string or loaded from a file using the `file` function. Supported private keys are unencrypted pem encoded RSA (PKCS#1), PKCS#8, DSA (OpenSSL), and ECDSA private keys. Mutually exclusive with `%[1]s`.", SchemaAttrSshPassword, SchemaAttrSshPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrSshPassword).String(),
			},
			DefaultFunc:      schemaEnvDefaultFunc(SchemaAttrSshPrivateKey, envPrefix, nil),
			ValidateDiagFunc: validate.PrivateKey(),
		},
		SchemaAttrSshCertificate: {
			Description: fmt.Sprintf("The ssh user certificate to authenticate with the remote ssh server. The certificate can be provided as text or loaded from a file using the `file` function. Expected format of the certificate is a base64 encoded OpenSSH public key (`authorized_keys` format). Must be used with in conjunction with `%[2]s`. Mutually exclusive with `%[1]s`.", SchemaAttrSshPassword, SchemaAttrSshPrivateKey),
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrSshPassword).String(),
			},
			DefaultFunc:      schemaEnvDefaultFunc(SchemaAttrSshCertificate, envPrefix, nil),
			ValidateDiagFunc: validate.AuthorizedKey(),
		},
		SchemaAttrSshHost: {
			Description: "The host of the remote ssh server to connect to.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSshHost, envPrefix, nil),
		},
		SchemaAttrSshHostKey: {
			Description:      "The public key or the CA certificate of the remote ssh host to verify the remote authenticity. Expected format of the host key is a base64 encoded OpenSSH public key (`authorized_keys` format).",
			Type:             schema.TypeString,
			Optional:         true,
			DefaultFunc:      schemaEnvDefaultFunc(SchemaAttrSshHostKey, envPrefix, nil),
			ValidateDiagFunc: validate.AuthorizedKey(),
		},
		SchemaAttrSshPort: {
			Description:  "The port of the remote ssh server to connect to. Defaults to `22`.",
			Type:         schema.TypeInt,
			Optional:     true,
			DefaultFunc:  schemaEnvDefaultFunc(SchemaAttrSshPort, envPrefix, 22),
			ValidateFunc: validation.IsPortNumber,
		},
		SchemaAttrSshTimeout: {
			Description: "Timeout of a single connection attempt. Should be provided as a string like `30s` or `5m`. Defaults to 30 seconds (`30s`).",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateDiagFunc: validate.All(
				validate.DurationAtLeast(1*time.Second),
				validate.DurationAtMost(60*time.Minute),
			),
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSshTimeout, envPrefix, "30s"),
		},
		SchemaAttrSshAgent: {
			Description: "If `true`, an ssh agent is used to to authenticate. Defaults to `false`.",
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSshAgent, envPrefix, false),
		},
		SchemaAttrSshAgentIdentity: {
			Description: "The preferred identity from the ssh agent for authentication. Expected format of an identity is a base64 encoded OpenSSH public key (`authorized_keys` format).",
			Type:        schema.TypeString,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrSshAgentIdentities).String(),
			},
			ValidateDiagFunc: validate.AuthorizedKey(),
			DefaultFunc:      schemaEnvDefaultFunc(SchemaAttrSshAgentIdentity, envPrefix, nil),
		},
		SchemaAttrSshAgentIdentities: {
			Description: "List of preferred identities from the ssh agent for authentication. Expected format of an identity is a base64 encoded OpenSSH public key (`authorized_keys` format).",
			Type:        schema.TypeList,
			Optional:    true,
			ConflictsWith: []string{
				attrPath.Extend(SchemaAttrSshAgentIdentity).String(),
			},
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: validate.AuthorizedKey(),
			},
		},
	}
}

func expandSchemaSsh(v interface{}) (*SchemaSsh, error) {
	d, err := expandListSingle(v)
	if err != nil {
		return nil, err
	}

	s := &SchemaSsh{
		User:            d[SchemaAttrSshUser].(string),
		Password:        d[SchemaAttrSshPassword].(string),
		PrivateKey:      d[SchemaAttrSshPrivateKey].(string),
		Certificate:     d[SchemaAttrSshCertificate].(string),
		Host:            d[SchemaAttrSshHost].(string),
		HostKey:         d[SchemaAttrSshHostKey].(string),
		Port:            d[SchemaAttrSshPort].(int),
		Agent:           d[SchemaAttrSshAgent].(bool),
		AgentIdentities: []string{},
	}

	if val, ok := d[SchemaAttrSshAgentIdentity].(string); ok && val != "" {
		// Single preferred identity
		s.AgentIdentities = []string{val}
	} else if vals, ok := d[SchemaAttrSshAgentIdentities].([]interface{}); ok && len(vals) > 0 {
		// Multiple preferred identities
		for _, val := range vals {
			s.AgentIdentities = append(s.AgentIdentities, val.(string))
		}
	}

	if timeoutStr := d[SchemaAttrSshTimeout].(string); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, err
		}
		s.Timeout = timeout
	}

	return s, nil
}
