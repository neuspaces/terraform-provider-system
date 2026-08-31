package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"path"
	"regexp"
	"sync/atomic"
	"testing"
)

var (
	testFolderId uint32
)

type testFolderConfig struct {
	folderName string
}

func newTestFolderConfig() testFolderConfig {
	id := atomic.AddUint32(&testFolderId, 1)

	return testFolderConfig{
		folderName: fmt.Sprintf("folder-%d", id),
	}
}

func TestAccFolder_create(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "755"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.test", "id", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.test", "path", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.test", "mode", "755"),
						resource.TestCheckResourceAttr("system_folder.test", "user", "root"),
						resource.TestCheckResourceAttr("system_folder.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_folder.test", "group", "root"),
						resource.TestCheckResourceAttr("system_folder.test", "gid", "0"),
					),
				},
			},
		})
	})
}

func TestAccFolder_update_mode(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "755"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.test", "id", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.test", "mode", "755"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "777"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.test", "id", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.test", "mode", "777"),
					),
				},
			},
		})
	})
}

func TestAccFolder_create_overwrite(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("existing", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "755"),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.existing", "id", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.existing", "mode", "755"),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("existing", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "755"),
						),
						testAccFolderBlock("test", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeString("mode", "700"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.test", "id", testRunFolderPath(target, testConfig.folderName)),
						resource.TestCheckResourceAttr("system_folder.test", "overwrite", "true"),
						resource.TestCheckResourceAttr("system_folder.test", "mode", "700"),
					),
				},
			},
		})
	})
}

func TestAccFolder_create_overwrite_fail_type_mismatch(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("existing", testRunFolderPath(target, testConfig.folderName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_file.existing", "id", testRunFolderPath(target, testConfig.folderName)),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFileBlock("existing", testRunFolderPath(target, testConfig.folderName)),
						testAccFolderBlock("test", testRunFolderPath(target, testConfig.folderName),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					ExpectError: regexp.MustCompile(`folder resource\s+folder path exists`),
				},
			},
		})
	})
}

func testRunFolderPath(target acctest.Target, p string) string {
	return path.Join(target.BasePath, p)
}

func testAccFolderBlock(name string, path string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("path", path),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_folder", name, resourceAttrs...)
}
