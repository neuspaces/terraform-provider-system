package provider_test

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv(acctest.EnvTfAcc) == "" {
		// Skip acceptance tests
		os.Exit(m.Run())
	}

	err := acctest.Initialize(m)
	if err != nil {
		msg := err.Error()
		if !strings.HasPrefix(msg, "[ERROR]") {
			msg = "[ERROR] " + strings.TrimLeft(msg, " ")
		}

		log.Print(msg)
		os.Exit(1)
	}
}

// TestProvider_InternalValidate validates the internal structure of the provider.Provider
func TestProvider_InternalValidate(t *testing.T) {
	providerFactory := provider.New("dev")
	p := providerFactory()

	err := p.InternalValidate()
	require.NoError(t, err)
}

// TestProvider_Configure_hostKeyPolicy verifies the v0.6.0 host key policy without a network connection:
// provider.Configure prepares the ssh client lazily and never dials.
func TestProvider_Configure_hostKeyPolicy(t *testing.T) {
	// The policy is decided by the attribute, which falls back to this environment variable.
	// Neutralize it so that the outcome does not depend on the environment of the test runner.
	t.Setenv(provider.SchemaEnvInsecureIgnoreHostKey, "")

	hostKey := strings.TrimSpace(string(testAccEd25519PrimaryPublicKey))

	configure := func(t *testing.T, raw map[string]interface{}) diag.Diagnostics {
		p := provider.New("dev")()
		stopCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := context.WithValue(context.Background(), schema.StopContextKey, stopCtx)
		return p.Configure(ctx, terraform.NewResourceConfigRaw(raw))
	}

	sshBlock := func(hostKey string) []interface{} {
		b := map[string]interface{}{"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret"}
		if hostKey != "" {
			b["host_key"] = hostKey
		}
		return []interface{}{b}
	}

	t.Run("ssh block without host_key is refused", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock("")})
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "SSH host key verification is not configured")
		require.Contains(t, diags[0].Detail, "insecure_ignore_host_key")
		require.Contains(t, diags[0].Detail, "ssh-keyscan")
	})

	t.Run("insecure_ignore_host_key true connects with a warning", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock(""), "insecure_ignore_host_key": true})
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, diag.Warning, diags[0].Severity)
		require.Contains(t, diags[0].Summary, "SSH host key verification is disabled")
	})

	t.Run("host_key set is silent", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock(hostKey)})
		require.Empty(t, diags)
	})

	t.Run("host_key set and insecure flag true is silent", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock(hostKey), "insecure_ignore_host_key": true})
		require.Empty(t, diags)
	})

	t.Run("proxy without host_key is refused", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{
			"ssh":   sshBlock(hostKey),
			"proxy": []interface{}{map[string]interface{}{"ssh": sshBlock("")}},
		})
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "SSH host key verification is not configured")
	})

	// The warning must name the place which is actually missing a host key. The remote has one here,
	// so a warning which talks about the `ssh` block would send the user to the wrong block.
	t.Run("proxy without host_key with insecure flag warns about the proxy", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{
			"ssh":                      sshBlock(hostKey),
			"proxy":                    []interface{}{map[string]interface{}{"ssh": sshBlock("")}},
			"insecure_ignore_host_key": true,
		})
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, diag.Warning, diags[0].Severity)
		require.Contains(t, diags[0].Summary, "SSH host key verification is disabled")
		require.Contains(t, diags[0].Detail, "`proxy.ssh`")
		require.Contains(t, diags[0].Detail, "bastion_host_key")
		require.NotContains(t, diags[0].Detail, "in the `ssh` block")
	})

	t.Run("connection block without host_key is refused", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"connection": []interface{}{map[string]interface{}{
			"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret",
		}}})
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "SSH host key verification is not configured")
		require.Contains(t, diags[0].Detail, "bastion_host_key")
	})

	// A `connection` block with a bastion is expanded into a proxy, so the host key of the bastion
	// is subject to the same policy as the host key of the remote
	t.Run("connection block with bastion without bastion_host_key is refused", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"connection": []interface{}{map[string]interface{}{
			"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret", "host_key": hostKey,
			"bastion_host": "127.0.0.2", "bastion_port": 22, "bastion_user": "root", "bastion_password": "secret",
		}}})
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "SSH host key verification is not configured")
		require.Contains(t, diags[0].Detail, "bastion_host_key")
	})

	t.Run("connection block with bastion and bastion_host_key is silent", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"connection": []interface{}{map[string]interface{}{
			"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret", "host_key": hostKey,
			"bastion_host": "127.0.0.2", "bastion_port": 22, "bastion_user": "root", "bastion_password": "secret",
			"bastion_host_key": hostKey,
		}}})
		require.Empty(t, diags)
	})

	// A single environment variable in a runner disables verification for every configuration without
	// leaving a trace in the configuration, so the diagnostic has to name the environment variable
	t.Run("insecure_ignore_host_key from environment names the environment variable", func(t *testing.T) {
		t.Setenv(provider.SchemaEnvInsecureIgnoreHostKey, "true")

		diags := configure(t, map[string]interface{}{"ssh": sshBlock("")})
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, diag.Warning, diags[0].Severity)
		require.Contains(t, diags[0].Summary, "SSH host key verification is disabled")
		require.Contains(t, diags[0].Summary, "TF_PROVIDER_SYSTEM_INSECURE_IGNORE_HOST_KEY")
		require.Contains(t, diags[0].Detail, "TF_PROVIDER_SYSTEM_INSECURE_IGNORE_HOST_KEY")
	})

	// By contrast, the attribute in the configuration is visible in review, so the diagnostic must
	// not blame the environment
	t.Run("insecure_ignore_host_key from configuration does not name the environment variable", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock(""), "insecure_ignore_host_key": true})
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.NotContains(t, diags[0].Summary, "TF_PROVIDER_SYSTEM_INSECURE_IGNORE_HOST_KEY")
		require.NotContains(t, diags[0].Detail, "TF_PROVIDER_SYSTEM_INSECURE_IGNORE_HOST_KEY")
	})
}

// TestProvider_Configure_hostKeyPolicy_rawConfigSource pins the behaviour of the production code path.
// (helper/schema.Provider).Configure copies the *terraform.ResourceConfig into the ResourceData so that
// GetRawConfig works, but terraform.NewResourceConfigRaw leaves CtyValue null, which is what the plain
// unit tests above exercise. Only the real plugin server populates CtyValue, so this test sets it by
// hand: it is the path which distinguishes an attribute set in the configuration from an attribute
// defaulted from the environment, and the path which would panic if GetAttr were called unguarded on a
// null object.
func TestProvider_Configure_hostKeyPolicy_rawConfigSource(t *testing.T) {
	t.Setenv(provider.SchemaEnvInsecureIgnoreHostKey, "")

	sshBlock := []interface{}{map[string]interface{}{"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret"}}

	configure := func(t *testing.T, raw map[string]interface{}, ctyValue cty.Value) diag.Diagnostics {
		p := provider.New("dev")()
		stopCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := context.WithValue(context.Background(), schema.StopContextKey, stopCtx)

		rc := terraform.NewResourceConfigRaw(raw)
		rc.CtyValue = ctyValue

		return p.Configure(ctx, rc)
	}

	t.Run("attribute present in raw config is not attributed to the environment", func(t *testing.T) {
		diags := configure(t,
			map[string]interface{}{"ssh": sshBlock, "insecure_ignore_host_key": true},
			cty.ObjectVal(map[string]cty.Value{provider.SchemaAttrInsecureIgnoreHostKey: cty.BoolVal(true)}),
		)
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Equal(t, "SSH host key verification is disabled", diags[0].Summary)
		require.NotContains(t, diags[0].Detail, provider.SchemaEnvInsecureIgnoreHostKey)
	})

	t.Run("attribute null in raw config is attributed to the environment", func(t *testing.T) {
		t.Setenv(provider.SchemaEnvInsecureIgnoreHostKey, "true")

		diags := configure(t,
			map[string]interface{}{"ssh": sshBlock},
			cty.ObjectVal(map[string]cty.Value{provider.SchemaAttrInsecureIgnoreHostKey: cty.NullVal(cty.Bool)}),
		)
		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		require.Contains(t, diags[0].Summary, provider.SchemaEnvInsecureIgnoreHostKey)
		require.Contains(t, diags[0].Detail, provider.SchemaEnvInsecureIgnoreHostKey)
	})

	t.Run("null raw config does not panic", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"ssh": sshBlock}, cty.NilVal)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "SSH host key verification is not configured")
	})
}
