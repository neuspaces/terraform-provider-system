package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
)

func TestAccDataFile_read_content(t *testing.T) {
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
						tfbuild.Data("system_file", "test",
							tfbuild.AttributeTraversal("path", tfbuild.TraversalResourceAttribute("system_file", "test", "path")),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("data.system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("data.system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("data.system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("data.system_file.test", "gid", "0"),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("data.system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
					),
				},
			},
		})
	})
}
