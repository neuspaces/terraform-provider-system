package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/sshclient"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	systemssh "github.com/neuspaces/terraform-provider-system/internal/system/ssh"
	"github.com/sethvargo/go-retry"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

const Name = "system"

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		return &schema.Provider{
			Schema:               providerSchema(),
			ResourcesMap:         providerResources(),
			DataSourcesMap:       providerDataSources(),
			ConfigureContextFunc: configure(),
		}
	}
}

type Provider struct {
	Config Schema
	System system.System
}

func init() {
	// Set descriptions to support markdown syntax
	schema.DescriptionKind = schema.StringMarkdown
}

func providerResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		resourceFileName:           resourceFile(),
		resourceFolderName:         resourceFolder(),
		resourceLinkName:           resourceLink(),
		resourceUserName:           resourceUser(),
		resourceGroupName:          resourceGroup(),
		resourceServiceOpenrcName:  resourceServiceOpenrc(),
		resourceServiceSystemdName: resourceServiceSystemd(),
		resourceSystemdUnitName:    resourceSystemdUnit(),
		resourcePackagesApkName:    resourcePackagesApk(),
		resourcePackagesAptName:    resourcePackagesApt(),
	}
}

func providerDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		dataReleaseName:  dataRelease(),
		dataIdentityName: dataIdentity(),
		dataCommandName:  dataCommand(),
		dataFileName:     dataFile(),
		dataFileMetaName: dataFileMeta(),
	}
}

func configure() schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Obtain stop context from parent context
		// terraform-plugin-sdk signals the provider stop by cancelling a *separate* context (stop context)
		// the stop context is contained as value with key schema.StopContextKey in ctx
		// https://github.com/hashicorp/terraform-plugin-sdk/blob/4681738a561387fb0b3aaa69aeb42231383634a0/helper/schema/grpc_provider.go#L554
		stopCtx, ok := schema.StopContext(ctx)
		if !ok {
			return nil, newInternalErrorDiagnostic("missing stop context in context of ConfigureContextFunc")
		}

		// Parse config
		c, err := expandProviderSchema(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Require ssh schema
		if c.Ssh == nil {
			return nil, newInternalUnexpectedTypeDiagnostic("*SchemaSsh", c.Ssh)
		}

		// Host key policy (v0.6.0): refuse unverified connections unless explicitly allowed
		diags := hostKeyPolicyDiagnostics(c)
		if diags.HasError() {
			return nil, diags
		}

		// Configure ssh client
		sshConnectOpts := sshConnectOptsFromSshSchema(*c.Ssh)

		if c.Proxy != nil && c.Proxy.Ssh != nil {
			// Connect via ssh proxy
			sshProxyConnectOpts := sshConnectOptsFromSshSchema(*c.Proxy.Ssh)

			sshProxyNetConnect := sshclient.Dial(sshclient.NewHostPortAddr(sshclient.Tcp, c.Proxy.Ssh.Host, uint16(c.Proxy.Ssh.Port)), c.Proxy.Ssh.Timeout)
			sshProxyConnectOpts = append(sshProxyConnectOpts, sshclient.Net(sshProxyNetConnect))
			sshProxyConnect, err := sshclient.Prepare(sshProxyConnectOpts...)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			sshNetConnect := sshclient.Proxy(sshclient.New(sshProxyConnect), sshclient.NewHostPortAddr(sshclient.Tcp, c.Ssh.Host, uint16(c.Ssh.Port)))

			sshConnectOpts = append(sshConnectOpts, sshclient.Net(sshNetConnect))

		} else {
			// Connect directly
			sshNetConnect := sshclient.Dial(sshclient.NewHostPortAddr(sshclient.Tcp, c.Ssh.Host, uint16(c.Ssh.Port)), c.Ssh.Timeout)

			sshConnectOpts = append(sshConnectOpts, sshclient.Net(sshNetConnect))
		}

		sshConnect, err := sshclient.Prepare(sshConnectOpts...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Retries
		if c.Retry {
			// retries are limited by maximum duration according to provider config with constant 1 second backoff
			retryM := sshclient.Retry(retry.WithMaxDuration(c.Timeout, retry.NewConstant(1*time.Second)))
			sshConnect = retryM(sshConnect)
		}

		// Break circuit when connection has failed once (after retries)
		cbM := sshclient.CircuitBreak()
		sshConnect = cbM(sshConnect)

		// Create ssh client
		sshClient := sshclient.New(sshConnect)

		// Configure system
		var sshSystemOpts []systemssh.SystemOption

		if c.Sudo {
			// Use sudo with shell /bin/sh
			sshSystemOpts = append(sshSystemOpts, systemssh.CommandMiddleware(cmd.SudoShMiddleware()))
		} else {
			// Use shell /bin/sh
			sshSystemOpts = append(sshSystemOpts, systemssh.CommandMiddleware(cmd.ShMiddleware()))
		}

		// Parallel sessions
		sshSystemOpts = append(sshSystemOpts, systemssh.Sessions(c.Parallel))

		s, err := systemssh.NewSystem(sshClient, sshSystemOpts...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Construct provider instance
		p := &Provider{
			Config: *c,
			System: s,
		}

		go func(ctx context.Context, s *systemssh.System) {
			// Wait for stop context cancelled
			<-stopCtx.Done()

			// Disconnect system
			_ = s.Close()
		}(ctx, s)

		return p, diags
	}
}

const (
	hostKeyMissingSummary  = "SSH host key verification is not configured"
	hostKeyDisabledSummary = "SSH host key verification is disabled"
)

// hostKeyPolicyDiagnostics enforces the host key policy introduced in v0.6.0: a connection to an ssh server
// without a configured host key is refused unless `insecure_ignore_host_key` is true, in which case a warning
// is reported. The policy applies to the remote and to the proxy connection alike.
func hostKeyPolicyDiagnostics(c *Schema) diag.Diagnostics {
	// Collect the places where a host key is expected but not configured, so that a diagnostic can
	// name them. The `connection` block is expanded into the same two places: `host_key` becomes the
	// host key of the remote, `bastion_host_key` the host key of the proxy.
	var missing []string

	if c.Ssh.HostKey == "" {
		missing = append(missing, fmt.Sprintf("`%s` in the `%s` block (or `%s` in the `%s` block)", SchemaAttrSshHostKey, SchemaAttrSsh, SchemaAttrConnectionHostKey, SchemaAttrConnection))
	}

	if c.Proxy != nil && c.Proxy.Ssh != nil && c.Proxy.Ssh.HostKey == "" {
		missing = append(missing, fmt.Sprintf("`%s` in the `%s.%s` block (or `%s` in the `%s` block)", SchemaAttrSshHostKey, SchemaAttrProxy, SchemaAttrSsh, SchemaAttrConnectionBastionHostKey, SchemaAttrConnection))
	}

	if len(missing) == 0 {
		return nil
	}

	if !c.InsecureIgnoreHostKey {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  hostKeyMissingSummary,
			Detail: fmt.Sprintf("The provider refuses to connect to an ssh server whose host key cannot be verified. "+
				"Set `%[1]s` in the `%[2]s` block (and in the `%[3]s.%[2]s` block when a proxy is used), or `%[1]s`/`%[4]s` in the `%[5]s` block, "+
				"to the public host key of the server, for example the output of `ssh-keyscan -t ed25519 <host>`. "+
				"Alternatively set `%[6]s = true` on the provider to restore the behavior of releases before v0.6.0, "+
				"which ignores the host key and exposes every connection to man-in-the-middle attacks.",
				SchemaAttrSshHostKey, SchemaAttrSsh, SchemaAttrProxy, SchemaAttrConnectionBastionHostKey, SchemaAttrConnection, SchemaAttrInsecureIgnoreHostKey),
		}}
	}

	missingList := strings.Join(missing, " and ")

	if c.InsecureIgnoreHostKeyFromEnv {
		// The configuration gives no hint that host key verification is disabled, so name the
		// environment variable which disabled it
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("%s by the environment variable %s", hostKeyDisabledSummary, SchemaEnvInsecureIgnoreHostKey),
			Detail: fmt.Sprintf("The environment variable `%[1]s` sets `%[2]s` to `true` and %[3]s is not set: "+
				"the identity of that ssh server is not verified. The provider configuration does not show this. "+
				"Unset `%[1]s` and set %[3]s to protect the connection against man-in-the-middle attacks.",
				SchemaEnvInsecureIgnoreHostKey, SchemaAttrInsecureIgnoreHostKey, missingList),
		}}
	}

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  hostKeyDisabledSummary,
		Detail: fmt.Sprintf("`%[1]s` is `true` and %[2]s is not set: the identity of that ssh server is not verified. "+
			"Set it to protect the connection against man-in-the-middle attacks.", SchemaAttrInsecureIgnoreHostKey, missingList),
	}}
}

func sshConnectOptsFromSshSchema(s SchemaSsh) []sshclient.ConnectOption {
	var sshConnectOpts []sshclient.ConnectOption

	// Address
	sshConnectOpts = append(sshConnectOpts, sshclient.Addr(sshclient.NewHostPortAddr(sshclient.Tcp, s.Host, uint16(s.Port))))

	// User
	sshConnectOpts = append(sshConnectOpts, sshclient.User(s.User))

	// Password
	if s.Password != "" {
		sshConnectOpts = append(sshConnectOpts, sshclient.Auth(sshclient.Password(s.Password)))
	}

	// Private key
	if s.PrivateKey != "" {
		sshConnectOpts = append(sshConnectOpts, sshclient.Auth(sshclient.PrivateKey(s.PrivateKey)))
	}

	// Certificate
	if s.Certificate != "" && s.PrivateKey != "" {
		sshConnectOpts = append(sshConnectOpts, sshclient.Auth(sshclient.Certificate(s.Certificate, s.PrivateKey)))
	}

	// Agent
	if s.Agent {
		var sshAuthMethod sshclient.AuthMethod

		if len(s.AgentIdentities) > 0 {
			// List of explicit identities
			sshAuthMethod = sshclient.AgentExplicitIdentities(s.AgentIdentities...)
		} else {
			// No explicit identities
			sshAuthMethod = sshclient.Agent()
		}

		sshConnectOpts = append(sshConnectOpts, sshclient.Auth(sshAuthMethod))

	}

	// Host key
	if s.HostKey != "" {
		sshConnectOpts = append(sshConnectOpts, sshclient.HostKey(sshclient.StaticHostKey(s.HostKey)))
	} else {
		sshConnectOpts = append(sshConnectOpts, sshclient.HostKeyCallback(ssh.InsecureIgnoreHostKey()))
	}

	return sshConnectOpts
}

func providerFromMeta(meta interface{}) (*Provider, diag.Diagnostics) {
	p, isProvider := meta.(*Provider)
	if !isProvider {
		return nil, newInternalUnexpectedTypeDiagnostic("*Provider", meta)
	}
	return p, nil
}
