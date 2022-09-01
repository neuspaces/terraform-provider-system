package acctest

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
)

type providerFactoryMap map[string]func() (*schema.Provider, error)

// ProviderFactories returns the provider factories used in the acceptance tests
func ProviderFactories() providerFactoryMap {
	return map[string]func() (*schema.Provider, error){
		provider.Name: func() (*schema.Provider, error) {
			return provider.New("dev")(), nil
		},
	}
}

func initProvider(ctx context.Context, tc ConfigTarget, factories providerFactoryMap) (*schema.Provider, error) {
	// Get provider factory from factory provider
	systemProviderFactory := factories[provider.Name]
	if systemProviderFactory == nil {
		return nil, fmt.Errorf("factory for provider %s not found", provider.Name)
	}

	// Create provider
	systemProvider, err := systemProviderFactory()
	if err != nil {
		return nil, err
	}

	// Prepare provider configuration
	defaultTargetCfg, err := tc.Configs.Get(DefaultTargetConfigId)
	if err != nil {
		return nil, err
	}

	providerCfg := ProviderResourceConfig(defaultTargetCfg)

	// Prepare separate stop context to mimic grpc provider server
	// cancel stop context if the parent context is done
	// https://github.com/hashicorp/terraform-plugin-sdk/blob/4681738a561387fb0b3aaa69aeb42231383634a0/helper/schema/grpc_provider.go#L554 and
	// https://github.com/hashicorp/terraform-plugin-sdk/blob/4681738a561387fb0b3aaa69aeb42231383634a0/helper/schema/grpc_provider.go#L57
	stopCtx, cancelStopCtx := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		cancelStopCtx()
	}()

	configureCtx := context.WithValue(ctx, schema.StopContextKey, stopCtx)

	// Configure
	diagErr := systemProvider.Configure(configureCtx, providerCfg)
	if diagErr != nil {
		return nil, fmt.Errorf("failed to configure provider %s: %+v", provider.Name, diagErr)
	}

	return systemProvider, nil
}
