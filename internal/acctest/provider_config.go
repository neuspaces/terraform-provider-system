package acctest

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
)

// ProviderConfigBlock is currently an alias for ProviderConfigBlockSshPasswordAuth
func ProviderConfigBlock(c ConfigTargetConfig) tfbuild.FileElement {
	sshAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString(provider.SchemaAttrSshHost, c.Ssh.Host),
		tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(c.Ssh.Port)),
		tfbuild.AttributeString(provider.SchemaAttrSshUser, c.Ssh.User),
	}

	if c.Ssh.Password != "" {
		sshAttrs = append(sshAttrs, tfbuild.AttributeString(provider.SchemaAttrSshPassword, c.Ssh.Password))
	}

	if c.Ssh.PrivateKey != "" {
		sshAttrs = append(sshAttrs, tfbuild.AttributeString(provider.SchemaAttrSshPrivateKey, c.Ssh.PrivateKey))
	}

	return tfbuild.Provider(provider.Name,
		tfbuild.InnerBlock(provider.SchemaAttrSsh, sshAttrs...),
	)
}

// ProviderConfigBlockSshPasswordAuth returns a provider configuration which uses ssh password authentication
func ProviderConfigBlockSshPasswordAuth(c ConfigTargetConfig) tfbuild.FileElement {
	return tfbuild.Provider(provider.Name,
		tfbuild.InnerBlock(provider.SchemaAttrSsh,
			tfbuild.AttributeString(provider.SchemaAttrSshHost, c.Ssh.Host),
			tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(c.Ssh.Port)),
			tfbuild.AttributeString(provider.SchemaAttrSshUser, c.Ssh.User),
			tfbuild.AttributeString(provider.SchemaAttrSshPassword, c.Ssh.Password),
		),
	)
}

func ProviderResourceConfig(c ConfigTargetConfig) *terraform.ResourceConfig {
	sshAttrs := map[string]interface{}{
		provider.SchemaAttrSshHost: c.Ssh.Host,
		provider.SchemaAttrSshPort: c.Ssh.Port,
		provider.SchemaAttrSshUser: c.Ssh.User,
	}

	if c.Ssh.Password != "" {
		sshAttrs[provider.SchemaAttrSshPassword] = c.Ssh.Password
	}

	if c.Ssh.PrivateKey != "" {
		sshAttrs[provider.SchemaAttrSshPrivateKey] = c.Ssh.PrivateKey
	}

	return terraform.NewResourceConfigRaw(map[string]interface{}{
		provider.SchemaAttrSsh: []interface{}{
			sshAttrs,
		},
	})
}
