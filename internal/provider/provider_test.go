package provider_test

import (
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
