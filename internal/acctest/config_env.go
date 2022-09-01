package acctest

import (
	"context"
	"github.com/sethvargo/go-envconfig"
)

// EnvConfig represents the acceptance test configuration which is provided by environment variables
type EnvConfig struct {
	// Targets is a list of target identifiers to include in the acceptance test run
	Targets []string `env:"TARGETS"`

	// ConfigPath is the path to the config yaml
	ConfigPath string `env:"CONFIG_PATH,default=acctest.yaml"`
}

// loadEnvConfig returns an EnvConfig which has been obtained from environment variables using an envconfig.Lookuper
func loadEnvConfig(ctx context.Context, lookuper envconfig.Lookuper) (*EnvConfig, error) {
	var p EnvConfig
	err := envconfig.ProcessWith(ctx, &p, lookuper)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
