package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
)

func TestAccDataFileMeta_read(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "hello world!"),
						),
						tfbuild.Data("system_file_meta", "test",
							tfbuild.AttributeTraversal("path", tfbuild.TraversalResourceAttribute("system_file", "test", "path")),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.system_file_meta.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "mode", "644"),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "user", "root"),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "uid", "0"),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "group", "root"),
						resource.TestCheckResourceAttr("data.system_file_meta.test", "gid", "0"),
					),
				},
			},
		})
	})
}
