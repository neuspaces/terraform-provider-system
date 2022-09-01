package provider_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"regexp"
	"testing"
)

// Test to install a single apk package which is not installed
//
// Preconditions:
// - Package `openssl` is not installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `openssl` is not installed after destroy
func TestAccPackagesApk_create_single(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
			},
		})
	})
}

// Test to install a single apk package which is not installed with a version constraint
//
// Preconditions:
// - Package `grep` is not installed
//
// Expected:
// - Package `grep=3.7-r0` is installed after create
// - Package `grep` is not installed after destroy
func TestAccPackagesApk_create_with_version(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=grep&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "grep"),
								tfbuild.AttributeString("version", "=3.7-r0"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.version", "=3.7-r0"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"grep":false}}`),
					),
				},
			},
		})
	})
}

// Test to install a single apk package which is already installed
//
// Preconditions:
// - Package `rsync` is installed
//
// Expected:
// - Package `rsync` is installed after create
// - Package `rsync` is installed after destroy
func TestAccPackagesApk_create_single_idempotent(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=rsync&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "rsync"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "rsync"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"rsync":true}}`),
					),
				},
			},
		})
	})
}

// Test to install multiple apk packages with mixed preconditions
//
// Preconditions:
// - Package `openssl` is not installed
// - Package `rsync` is installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `openssl` is not installed after destroy
// - Package `rsync` is installed after create
// - Package `rsync` is installed after destroy
func TestAccPackagesApk_multiple(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://pkgs.alpinelinux.org/packages?name=rsync&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "rsync"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.1.name", "rsync"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.1.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"openssl":false,"rsync":true}}`),
					),
				},
			},
		})
	})
}

// Test to install a single apk package and add a second apk package
//
// Preconditions:
// - Package `openssl` is not installed
// - Package `grep` is not installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `grep` is installed after update
// - Package `openssl` is not installed after destroy
// - Package `grep` is not installed after destroy
func TestAccPackagesApk_update_add_package(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://pkgs.alpinelinux.org/packages?name=grep&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "grep"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "grep"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.1.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"grep":false,"openssl":false}}`),
					),
				},
			},
		})
	})
}

// Test to install multiple apk packages and subsequently remove the second apk package
//
// Preconditions:
// - Package `openssl` is not installed
// - Package `grep` is not installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `grep` is installed after create
// - Package `grep` is not installed after update
// - Package `openssl` is not installed after destroy
// - Package `grep` is not installed after destroy
func TestAccPackagesApk_update_remove_package(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://pkgs.alpinelinux.org/packages?name=grep&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "grep"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "grep"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.1.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.1.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"grep":false,"openssl":false}}`),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apk.test"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apk.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttrSet("system_packages_apk.test", "package.0.versions.0.available"),
						provider.TestCheckResourceAttrBase64("system_packages_apk.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
			},
		})
	})
}

func TestAccPackagesApk_unavailable(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageApkBlock("test",
							// https://pkgs.alpinelinux.org/packages?name=openssl&branch=v3.15
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					)),
					ExpectError: regexp.MustCompile(`apk not available`),
				},
			},
		})
	})
}

func testAccPackageApkBlock(name string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	return tfbuild.Resource("system_packages_apk", name, attrs...)
}
