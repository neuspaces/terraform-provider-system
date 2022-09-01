package provider_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"testing"
)

func TestAccDataIdentity_default(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		targetConfig := target.Configs.Default()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(targetConfig),
						testAccDataIdentityBlock("test"),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "data.system_identity.test"),
						resource.TestCheckResourceAttrSet("data.system_identity.test", "id"),
						resource.TestCheckResourceAttr("data.system_identity.test", "user", targetConfig.Ssh.User),
						resource.TestCheckResourceAttr("data.system_identity.test", "uid", "0"),
						resource.TestCheckResourceAttr("data.system_identity.test", "group", "root"),
						resource.TestCheckResourceAttr("data.system_identity.test", "gid", "0"),
					),
				},
			},
		})
	})
}

func testAccDataIdentityBlock(name string) tfbuild.FileElement {
	return tfbuild.Data("system_identity", name)
}
