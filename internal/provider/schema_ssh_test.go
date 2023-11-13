package provider_test

import (
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/sshagent"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"regexp"
	"testing"
)

var (
	//go:embed test/ed25519-primary
	testAccEd25519PrimaryPrivateKey []byte

	//go:embed test/ed25519-primary.pub
	testAccEd25519PrimaryPublicKey []byte

	//go:embed test/ed25519-secondary
	testAccEd25519SecondaryPrivateKey []byte

	//go:embed test/ed25519-secondary.pub
	testAccEd25519SecondaryPublicKey []byte

	//go:embed test/ecdsa-primary
	testAccEcdsaPrimaryPrivateKey []byte

	//go:embed test/ecdsa-primary.pub
	testAccEcdsaPrimaryPublicKey []byte

	//go:embed test/ecdsa-secondary
	testAccEcdsaSecondaryPrivateKey []byte

	//go:embed test/ecdsa-secondary.pub
	testAccEcdsaSecondaryPublicKey []byte
)

func getTargetConfigOrSkip(t *testing.T, target acctest.Target, targetConfigId string) acctest.ConfigTargetConfig {
	targetConfig, err := target.Configs.Get(targetConfigId)
	if err != nil {
		t.Skip(fmt.Sprintf("target %q is missing config %q", target.Id, targetConfigId))
	}

	return targetConfig
}

func testAccProviderConnectTestExpectConnect(t *testing.T, targetConfig acctest.ConfigTargetConfig, providerConfig tfbuild.FileElement) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConnectTestConfig(providerConfig),
				Check:  testAccProviderConnectTestCheckFunc(t, targetConfig),
			},
		},
	})
}

func testAccProviderConnectTestExpectError(t *testing.T, providerConfig tfbuild.FileElement, errorRegexp *regexp.Regexp) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConnectTestConfig(providerConfig),
				ExpectError: errorRegexp,
			},
		},
	})
}

func testAccProviderConnectTestConfig(providerConfig tfbuild.FileElement) string {
	return tfbuild.FileString(tfbuild.File(
		providerConfig,
		testAccDataIdentityBlock("test"),
	))
}

func testAccProviderConnectTestCheckFunc(t *testing.T, targetConfig acctest.ConfigTargetConfig) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		provider.TestLogResourceAttr(t, "data.system_identity.test"),
		resource.TestCheckResourceAttr("data.system_identity.test", "user", targetConfig.Ssh.User),
	)
}

func TestAccProviderConnect_SshPassword(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
			)

			testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
		})
	})

	t.Run("wrong password", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, "wrong!"),
				),
			)

			testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`ssh: handshake failed: ssh: unable to authenticate`)))
		})
	})

}

func TestAccProviderConnect_SshPrivateKey(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPrivateKey, targetConfig.Ssh.PrivateKey),
				),
			)

			testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
		})
	})

	t.Run("unauthorized private key", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			unauthorizedPrivateKey := string(testAccEd25519PrimaryPrivateKey)

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPrivateKey, unauthorizedPrivateKey),
				),
			)

			testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain`)))
		})
	})
}

func TestAccProviderConnect_SshAgent(t *testing.T) {
	t.Run("no explicit agent identities", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			agentPrivateKey := sshagent.PrivateKey(targetConfig.Ssh.PrivateKey)

			sshagent.NewTestServer(t, agentPrivateKey).Use(t, func(t *testing.T) {
				providerConfig := tfbuild.Provider(provider.Name,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
						tfbuild.AttributeBool(provider.SchemaAttrSshAgent, true),
					),
				)

				testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
			})
		})
	})

	t.Run("no explicit agent identities with empty agent", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			sshagent.NewTestServer(t).Use(t, func(t *testing.T) {
				providerConfig := tfbuild.Provider(provider.Name,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
						tfbuild.AttributeBool(provider.SchemaAttrSshAgent, true),
					),
				)

				testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`ssh: handshake failed: ssh: unable to authenticate`)))
			})
		})
	})

	t.Run("single explicit agent identity with matching identity", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			agentPrivateKey := sshagent.PrivateKey(targetConfig.Ssh.PrivateKey)

			sshagent.NewTestServer(t, agentPrivateKey).Use(t, func(t *testing.T) {
				providerConfig := tfbuild.Provider(provider.Name,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
						tfbuild.AttributeBool(provider.SchemaAttrSshAgent, true),
						tfbuild.Attribute(provider.SchemaAttrSshAgentIdentities, tfbuild.StringList(
							targetConfig.Ssh.PublicKey,
						)),
					),
				)

				testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
			})
		})
	})

	t.Run("single agent identity without matching identity", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			agentPrivateKey := sshagent.PrivateKey(targetConfig.Ssh.PrivateKey)

			sshagent.NewTestServer(t, agentPrivateKey).Use(t, func(t *testing.T) {
				providerConfig := tfbuild.Provider(provider.Name,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
						tfbuild.AttributeBool(provider.SchemaAttrSshAgent, true),
						tfbuild.Attribute(provider.SchemaAttrSshAgentIdentities, tfbuild.StringList(
							string(testAccEd25519PrimaryPublicKey),
						)),
					),
				)

				testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`ssh: handshake failed: sshclient: agent does not provide any explicit identity`)))
			})
		})
	})

	t.Run("multiple agent identities with matching identity", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-private-key")

			agentPrivateKey := sshagent.PrivateKey(targetConfig.Ssh.PrivateKey)

			sshagent.NewTestServer(t, agentPrivateKey).Use(t, func(t *testing.T) {
				providerConfig := tfbuild.Provider(provider.Name,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
						tfbuild.AttributeBool(provider.SchemaAttrSshAgent, true),
						tfbuild.Attribute(provider.SchemaAttrSshAgentIdentities, tfbuild.StringList(
							string(testAccEd25519PrimaryPublicKey),
							targetConfig.Ssh.PublicKey,
						)),
					),
				)

				testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
			})
		})
	})
}

func TestAccProviderConnect_SshHostKey(t *testing.T) {
	t.Run("no host key", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
			)

			testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
		})
	})

	t.Run("valid host key", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshHostKey, targetConfig.Ssh.HostKey),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
			)

			testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
		})
	})

	t.Run("invalid host key format", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			// openssl rand 51 | base64
			invalidHostKey := "i+8cl1wtm8Yq8r3er3jsEsDBFX4oaGbBV/Hq3N2mKb+hN55jbUCWbCkZNhk5uNa4OsTn"

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshHostKey, invalidHostKey),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
			)

			testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`invalid authorized key format`)))
		})
	})

	t.Run("host key mismatch", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshHostKey, string(testAccEcdsaPrimaryPublicKey)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
			)

			testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(regexp.QuoteMeta(`ssh: handshake failed: ssh: host key mismatch`)))
		})
	})
}

func TestAccProviderSudo(t *testing.T) {
	t.Run("privileged user should become root", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-privileged")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
				tfbuild.AttributeBool(provider.SchemaAttrSudo, true),
			)

			resource.Test(t, resource.TestCase{
				ProviderFactories: acctest.ProviderFactories(),
				Steps: []resource.TestStep{
					{
						Config: testAccProviderConnectTestConfig(providerConfig),
						Check: resource.ComposeTestCheckFunc(
							provider.TestLogResourceAttr(t, "data.system_identity.test"),
							resource.TestCheckResourceAttr("data.system_identity.test", "user", "root"),
						),
					},
				},
			})
		})
	})

	t.Run("unprivileged user should not become root", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			targetConfig := getTargetConfigOrSkip(t, target, "auth-unprivileged")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, targetConfig.Ssh.Host),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(targetConfig.Ssh.Port)),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
				tfbuild.AttributeBool(provider.SchemaAttrSudo, true),
			)

			testAccProviderConnectTestExpectError(t, providerConfig, regexp.MustCompile(`info resource\s+unexpected error`))
		})
	})
}

func TestAccProviderConnect_SshProxy(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
			t.Parallel()

			proxyTargetConfig := getTargetConfigOrSkip(t, target, "auth-unprivileged")
			targetConfig := getTargetConfigOrSkip(t, target, "auth-password")

			providerConfig := tfbuild.Provider(provider.Name,
				tfbuild.InnerBlock(provider.SchemaAttrSsh,
					tfbuild.AttributeString(provider.SchemaAttrSshHost, "localhost"),
					tfbuild.AttributeInt(provider.SchemaAttrSshPort, 22),
					tfbuild.AttributeString(provider.SchemaAttrSshUser, targetConfig.Ssh.User),
					tfbuild.AttributeString(provider.SchemaAttrSshPassword, targetConfig.Ssh.Password),
				),
				tfbuild.InnerBlock(provider.SchemaAttrProxy,
					tfbuild.InnerBlock(provider.SchemaAttrSsh,
						tfbuild.AttributeString(provider.SchemaAttrSshHost, proxyTargetConfig.Ssh.Host),
						tfbuild.AttributeInt(provider.SchemaAttrSshPort, int64(proxyTargetConfig.Ssh.Port)),
						tfbuild.AttributeString(provider.SchemaAttrSshUser, proxyTargetConfig.Ssh.User),
						tfbuild.AttributeString(provider.SchemaAttrSshPassword, proxyTargetConfig.Ssh.Password),
					),
				),
			)

			testAccProviderConnectTestExpectConnect(t, targetConfig, providerConfig)
		})
	})
}
