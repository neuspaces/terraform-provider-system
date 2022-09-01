package acctest

const (
	// EnvPrefix is the prefix for all environment variables which are used to configure the acceptance tests
	// EnvPrefix is intentionally *different* from provider.Schema because different acceptance tests may test different provider configurations
	EnvPrefix = "TF_ACC_PROVIDER_SYSTEM_"
)

const (
	EnvTfProviderSystemTargets = EnvPrefix + "TARGETS"

	EnvTfProviderSystemConfigPath = EnvPrefix + "CONFIG_PATH"
)

// Environment variables defined by the Terraform acceptance test framework
// https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests#environment-variables
const (
	// EnvTfAcc refers to the TF_ACC environment variable
	EnvTfAcc = "TF_ACC"

	// EnvTfLog refers to the TF_LOG environment variable
	// https://www.terraform.io/plugin/log/managing#enable-logging
	EnvTfLog = "TF_LOG"

	// EnvTfAccTerraformPath refers to the TF_ACC_TERRAFORM_PATH environment variable
	EnvTfAccTerraformPath = "TF_ACC_TERRAFORM_PATH"
)
