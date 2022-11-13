package provider_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"testing"
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
						testAccDataFileBlock("test", testRunFilePath(target, testConfig.fileName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test", "gid", "0"),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),

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

func TestAccFile_read_content_sensitive(t *testing.T) {
	testConfig := newTestFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("test_sensitive", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content_sensitive", "hello s3cr3t!"),
						),
						testAccDataFileBlock("test_sensitive", testRunFilePath(target, testConfig.fileName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.test_sensitive", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "user", "root"),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "group", "root"),
						resource.TestCheckResourceAttr("system_file.test_sensitive", "gid", "0"),
						// echo -n 'hello s3cr3t!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.test_sensitive", "md5sum", "su6hNsFeImKJgWRpqtfzZw=="),

						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "mode", "644"),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "user", "root"),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "uid", "0"),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "group", "root"),
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "gid", "0"),
						// echo -n 'hello s3cr3t!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("data.system_file.test_sensitive", "md5sum", "su6hNsFeImKJgWRpqtfzZw=="),
					),
				},
			},
		})
	})
}

func testAccDataFileBlock(name string, path string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("path", path),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Data("system_file", name, resourceAttrs...)
}
