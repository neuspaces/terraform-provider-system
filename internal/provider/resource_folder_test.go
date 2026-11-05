package provider_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/client"
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

func TestAccFolder_overwrite_existing(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		folderPath := testRunFolderPath(target, testConfig.folderName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						err := client.NewFolderClient(target.Provider.System).Create(context.Background(), client.Folder{
							Path: folderPath, Mode: 0o755, Uid: -1, Gid: -1,
						})
						if err != nil {
							t.Fatalf("pre-create folder %s: %v", folderPath, err)
						}
					},
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", folderPath,
							tfbuild.AttributeString("mode", "700"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_folder.test", "id", folderPath),
						resource.TestCheckResourceAttr("system_folder.test", "overwrite", "true"),
						resource.TestCheckResourceAttr("system_folder.test", "mode", "700"),
					),
				},
			},
		})
	})
}

func TestAccFolder_overwrite_fail_type_mismatch(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		folderPath := testRunFolderPath(target, testConfig.folderName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						err := client.NewFileClient(target.Provider.System).Create(context.Background(), client.File{
							Path: folderPath, Uid: -1, Gid: -1,
						})
						if err != nil {
							t.Fatalf("pre-create file %s: %v", folderPath, err)
						}
					},
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", folderPath,
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					ExpectError: regexp.MustCompile(`folder resource\s+folder path exists`),
				},
			},
		})
	})
}

func TestAccFolder_overwrite_fail_symlink(t *testing.T) {
	testConfig := newTestFolderConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		folderPath := testRunFolderPath(target, testConfig.folderName)
		linkTargetPath := folderPath + ".target"

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						// `path` is a symbolic link to a folder elsewhere; `[ -e ]` and `[ -d ]` follow the link,
						// so without an explicit check for a link the permissions of the link target would be changed
						if _, err := testAccFileRemoteSh(target, fmt.Sprintf(`mkdir -m 755 '%[1]s' && touch '%[1]s/keep' && ln -s '%[1]s' '%[2]s'`, linkTargetPath, folderPath)); err != nil {
							t.Fatalf("pre-create symlink %s: %v", folderPath, err)
						}
					},
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccFolderBlock("test", folderPath,
							tfbuild.AttributeString("mode", "700"),
							tfbuild.AttributeBool("overwrite", true),
						),
					)),
					ExpectError: regexp.MustCompile(`folder resource\s+folder path exists`),
				},
			},
		})

		// the link target must be untouched: neither its permissions nor its content may be changed through the link
		out, err := testAccFileRemoteSh(target, fmt.Sprintf(`stat -c %%a '%[1]s'; ls '%[1]s'`, linkTargetPath))
		if err != nil {
			t.Error(err)
		} else if out != "755\nkeep" {
			t.Errorf("expected link target %s to keep mode 755 and its content, got: %q", linkTargetPath, out)
		}
	})
}
