package provider_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"testing"
)

func TestAccDataRelease_default(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccDataReleaseBlock("test"),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "data.system_release.test"),
						resource.TestCheckResourceAttrSet("data.system_release.test", "id"),
						resource.TestCheckResourceAttr("data.system_release.test", "name", target.Os.Name),
						resource.TestCheckResourceAttr("data.system_release.test", "vendor", target.Os.Vendor),
						resource.TestCheckResourceAttr("data.system_release.test", "version", target.Os.Version),
						resource.TestCheckResourceAttr("data.system_release.test", "release", target.Os.Release),
					),
				},
			},
		})
	})
}

func testAccDataReleaseBlock(name string) tfbuild.FileElement {
	return tfbuild.Data("system_release", name)
}
