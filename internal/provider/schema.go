package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"time"
)

// Schema is a struct to represent the configuration of the provider
type Schema struct {
	Ssh *SchemaSsh

	Proxy *SchemaProxy

	Parallel int
	Timeout  time.Duration
	Retry    bool

	Sudo bool
}

// expandProviderSchema returns a Schema from schema.ResourceData of the provider configuration
func expandProviderSchema(d *schema.ResourceData) (*Schema, error) {
	s := &Schema{}

	if sshV, sshOk := d.GetOk(SchemaAttrSsh); sshOk {
		// Standard configuration with `ssh` block
		// Recommended configuration using `ssh` block and optional `proxy` block
		schemaSsh, err := expandSchemaSsh(sshV)
		if err != nil {
			return nil, err
		}
		s.Ssh = schemaSsh

		// Optional proxy
		if proxySshV, proxySshOk := d.GetOk(newAttrPath(SchemaAttrProxy, "0", SchemaAttrSsh).String()); proxySshOk {
			proxySchemaSsh, err := expandSchemaSsh(proxySshV)
			if err != nil {
				return nil, err
			}

			s.Proxy = &SchemaProxy{
				Ssh: proxySchemaSsh,
			}
		}
	} else if connectionV, connectionOk := d.GetOk(SchemaAttrConnection); connectionOk {
		// Compatible configuration using `connection` block
		// Users may configure the provider using `connection` block which equals the
		// `connection` block in a Terraform provisioner configuration
		// https://www.terraform.io/language/resources/provisioners/connection
		schemaSsh, bastionSchemaSsh, err := expandSchemaConnection(connectionV)
		if err != nil {
			return nil, err
		}

		s.Ssh = schemaSsh

		if bastionSchemaSsh != nil {
			s.Proxy = &SchemaProxy{
				Ssh: bastionSchemaSsh,
			}
		}
	} else {
		// TODO support configuration from environment variables if neither `ssh` nor `connection` is configured
		return nil, fmt.Errorf("provider configuration requires either one of the following blocks: %s, %s", SchemaAttrSsh, SchemaAttrConnection)
	}

	// Other
	s.Parallel = d.Get(SchemaAttrParallel).(int)
	s.Retry = d.Get(SchemaAttrRetry).(bool)

	if timeoutStr := d.Get(SchemaAttrTimeout).(string); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, err
		}
		s.Timeout = timeout
	}

	s.Sudo = d.Get(SchemaAttrSudo).(bool)

	return s, nil
}

const (
	SchemaEnvPrefix = "TF_PROVIDER_SYSTEM_"
)

const (
	SchemaAttrConnection = "connection"

	SchemaAttrSsh   = "ssh"
	SchemaAttrProxy = "proxy"

	SchemaAttrParallel = "parallel"
	SchemaAttrTimeout  = "timeout"
	SchemaAttrRetry    = "retry"

	SchemaAttrShell = "shell"
	SchemaAttrSudo  = "sudo"
)

// providerSchema returns the provider schema
func providerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		SchemaAttrConnection: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			ConflictsWith: []string{
				SchemaAttrSsh,
				SchemaAttrProxy,
			},
			Elem: &schema.Resource{
				Schema: providerSchemaConnection(newAttrPath(SchemaAttrConnection, "0")),
			},
		},
		SchemaAttrSsh: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			ConflictsWith: []string{
				SchemaAttrConnection,
			},
			Elem: &schema.Resource{
				Schema: providerSchemaSsh(newAttrPath(SchemaAttrSsh, "0"), SchemaEnvPrefix+"SSH_"),
			},
		},
		SchemaAttrProxy: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			RequiredWith: []string{
				SchemaAttrSsh,
			},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					SchemaAttrSsh: {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: providerSchemaSsh(newAttrPath(SchemaAttrProxy, "0", SchemaAttrSsh, "0"), SchemaEnvPrefix+"PROXY_SSH_"),
						},
					},
				},
			},
		},
		SchemaAttrParallel: {
			Description:  "Maximum number of concurrent ssh connections to the remote. Increase the number of connections to parallelize interaction with the remote. Set to `0` to not limit the number of concurrent connections. Defaults to `1`.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 256),
			DefaultFunc:  schemaEnvDefaultFunc(SchemaAttrParallel, SchemaEnvPrefix, 1),
		},
		SchemaAttrTimeout: {
			Description: "Timeout for the connection to the remote to become available. This timeout include multiple connection attempts if retires are enabled. Provided as a duration string like `30s` or `5m`. Defaults to `5m`.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateDiagFunc: validate.All(
				validate.DurationAtLeast(1*time.Second),
				validate.DurationAtMost(60*time.Minute),
			),
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrTimeout, SchemaEnvPrefix, "5m"),
		},
		SchemaAttrRetry: {
			Description: "If `true`, the provider retries failed connection attempts to the remote within the configured timeout. A constant backoff of 1s is planned between failed connection attempts. Defaults to `true`.",
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrRetry, SchemaEnvPrefix, true),
		},
		SchemaAttrSudo: {
			Description: "If `true`, commands are executed on the remote using `sudo` by default. Enable `sudo` to connect to the remote with an unprivileged used and execute commands as root. As a prerequisite `sudo` must be installed and configured on the remote system. The `user` must be able to run `sudo` without password (`NOPASSWD`). Defaults to `false`.",
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schemaEnvDefaultFunc(SchemaAttrSudo, SchemaEnvPrefix, false),
		},
	}
}
