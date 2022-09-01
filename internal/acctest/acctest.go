package acctest

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/lib/contains"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"github.com/sethvargo/go-envconfig"
	"path"
	"testing"
)

type AccTest struct {
	// Id contains the unique id of the test run
	Id string

	Config Config

	Targets Targets
}

func Initialize(m *testing.M) error {
	var err error

	ctx, ctxDone := context.WithCancel(context.Background())
	defer func() {
		// Finish context at the end of tests
		ctxDone()
	}()

	// Load configuration from environment variables
	envCfg, err := loadEnvConfig(ctx, envconfig.PrefixLookuper(EnvPrefix, envconfig.OsLookuper()))
	if err != nil {
		return fmt.Errorf("invalid acceptance test configuration: %w", err)
	}
	if envCfg == nil {
		return fmt.Errorf("expected acceptance test configuration, got nil")
	}

	// Load configuration from config yaml
	cfgYaml, err := fileReadAll(envCfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config yaml: %w", err)
	}

	cfg, err := loadConfig(cfgYaml)
	if err != nil {
		return err
	}

	// Generate test run id (8 character hex string)
	testRunId := acctest.RandStringFromCharSet(8, "abcdef012346789")

	// Initialize AccTest
	current = AccTest{
		Id:      testRunId,
		Targets: []Target{},
	}

	// Initialize targets
	for targetId, targetCfg := range cfg.Targets {
		// Filter for explicitly enabled targets
		// Do not filter of explicitly enabled targets have not been provided
		if envCfg.Targets != nil && !contains.String(envCfg.Targets, targetId) {
			continue
		}

		// Initialize provider for each target
		providerSchema, err := initProvider(ctx, targetCfg, ProviderFactories())
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to initialize or configure %s provider: %w", provider.Name, err)
		}

		providerMeta, ok := providerSchema.Meta().(*provider.Provider)
		if !ok {
			return fmt.Errorf("[ERROR] Unexpected result as %s provider meta", provider.Name)
		}

		target := Target{
			ConfigTarget: targetCfg,
			Id:           targetId,
			BasePath:     path.Join("/root", "test"+testRunId),
			Prefix:       "test" + testRunId,
			Provider:     providerMeta,
		}

		// Setup of test targets
		err = setupTarget(ctx, &target)
		if err != nil {
			return err
		}

		current.Targets = append(current.Targets, target)
	}

	// Tear down of test targets
	defer func() {
		for _, target := range current.Targets {
			err := teardownTarget(ctx, &target)
			if err != nil {
				panic(err)
			}
		}
	}()

	// Run test with terraform provider sdk
	resource.TestMain(m)

	return nil
}

func setupTarget(ctx context.Context, target *Target) error {
	var err error

	// Create base path for tests which involve the file system
	err = client.NewFolderClient(target.Provider.System).Create(ctx, client.Folder{
		Path: target.BasePath,
	})
	if err != nil {
		return fmt.Errorf("[ERROR] Error creating test base path on system: %w", err)
	}

	return nil
}

func teardownTarget(ctx context.Context, target *Target) error {
	var err error

	err = client.NewFolderClient(target.Provider.System).Delete(ctx, target.BasePath)
	if err != nil {
		return fmt.Errorf("[ERROR] Error deleting test base path on system: %w", err)
	}

	return err
}
