package test

import "strings"

// ConcatTestConfig concats one or more Terraform configurations
func ConcatTestConfig(configs ...string) string {
	var trimmedConfigs []string
	for _, c := range configs {
		trimmedConfigs = append(trimmedConfigs, strings.TrimSpace(c))
	}

	return strings.TrimSpace(strings.Join(configs, "\n\n"))
}
