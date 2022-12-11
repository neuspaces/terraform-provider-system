package provider_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
)

func newTestDataFileConfig() testFileConfig {
	id := atomic.AddUint32(&testFileId, 1)

	return testFileConfig{
		userName: fmt.Sprintf("user%d", id),
		fileName: fmt.Sprintf("file-%d", id),
	}
}

func TestAccDataFile_read_content(t *testing.T) {
	testConfig := newTestDataFileConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("datatest", testRunFilePath(target, testConfig.fileName),
							tfbuild.AttributeString("mode", "644"),
							tfbuild.AttributeString("content", "hello world!"),
						),
						tfbuild.Data("system_file", "datatest",
							tfbuild.AttributeTraversal("path", tfbuild.TraversalResourceAttribute("system_file", "datatest", "path")),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.datatest", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.datatest", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("system_file.datatest", "mode", "644"),
						resource.TestCheckResourceAttr("system_file.datatest", "user", "root"),
						resource.TestCheckResourceAttr("system_file.datatest", "uid", "0"),
						resource.TestCheckResourceAttr("system_file.datatest", "group", "root"),
						resource.TestCheckResourceAttr("system_file.datatest", "gid", "0"),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("system_file.datatest", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),

						resource.TestCheckResourceAttr("data.system_file.datatest", "id", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.datatest", "path", testRunFilePath(target, testConfig.fileName)),
						resource.TestCheckResourceAttr("data.system_file.datatest", "mode", "644"),
						resource.TestCheckResourceAttr("data.system_file.datatest", "user", "root"),
						resource.TestCheckResourceAttr("data.system_file.datatest", "uid", "0"),
						resource.TestCheckResourceAttr("data.system_file.datatest", "group", "root"),
						resource.TestCheckResourceAttr("data.system_file.datatest", "gid", "0"),
						// echo -n 'hello world!' | openssl dgst -binary -md5 | openssl base64
						resource.TestCheckResourceAttr("data.system_file.datatest", "md5sum", "/D/5joxqDTCH1RXARz+Gdw=="),
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
