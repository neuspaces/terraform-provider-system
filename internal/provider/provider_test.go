package provider_test

import (
	"context"
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
	})

	t.Run("connection block without host_key is refused", func(t *testing.T) {
		diags := configure(t, map[string]interface{}{"connection": []interface{}{map[string]interface{}{
			"host": "127.0.0.1", "port": 22, "user": "root", "password": "secret",
		}}})
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Detail, "bastion_host_key")
	})
}
