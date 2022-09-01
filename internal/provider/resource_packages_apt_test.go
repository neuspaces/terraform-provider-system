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

// Test to install a single apt package which is not installed
//
// Preconditions:
// - Package `openssl` is not installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `openssl` is not installed after destroy
func TestAccPackagesApt_create_single(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
			},
		})
	})
}

// Test to install a single apt package which is already installed
//
// Preconditions:
// - Package `grep` is installed
//
// Expected:
// - Package `grep` is installed after create
// - Package `grep` is installed after destroy
func TestAccPackagesApt_create_single_idempotent(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/grep
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "grep"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "grep"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"grep":true}}`),
					),
				},
			},
		})
	})
}

// Test to install multiple apt packages with mixed preconditions
//
// Preconditions:
// - Package `openssl` is not installed
// - Package `grep` is installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `openssl` is not installed after destroy
// - Package `grep` is installed after create
// - Package `grep` is installed after destroy
func TestAccPackagesApt_multiple(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://packages.debian.org/de/bullseye/grep
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "grep"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "grep"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.name", "openssl"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.versions.#", "1"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"grep":true,"openssl":false}}`),
					),
				},
			},
		})
	})
}

// Test to install a single apt package and add a second apt package
//
// Preconditions:
// - Package `openssl` is not installed
// - Package `unzip` is not installed
//
// Expected:
// - Package `openssl` is installed after create
// - Package `unzip` is installed after update
// - Package `openssl` is not installed after destroy
// - Package `unzip` is not installed after destroy
func TestAccPackagesApt_update_add_package(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://packages.debian.org/bullseye/unzip
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "unzip"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.name", "unzip"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"openssl":false,"unzip":false}}`),
					),
				},
			},
		})
	})
}

func TestAccPackagesApt_update_remove_package(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
							// https://packages.debian.org/bullseye/unzip
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "unzip"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "2"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.name", "unzip"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.1.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.1.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"openssl":false,"unzip":false}}`),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_packages_apt.test"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "id"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.#", "1"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.name", "openssl"),
						resource.TestCheckResourceAttrSet("system_packages_apt.test", "package.0.versions.0.installed"),
						resource.TestCheckResourceAttr("system_packages_apt.test", "package.0.versions.0.available", ""),
						provider.TestCheckResourceAttrBase64("system_packages_apt.test", "internal", `{"pre_installed":{"openssl":false}}`),
					),
				},
			},
		})
	})
}

func TestAccPackagesApt_unavailable(t *testing.T) {
	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSEquals(t, target, osrelease.DebianId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccPackageAptBlock("test",
							// https://packages.debian.org/de/bullseye/openssl
							tfbuild.InnerBlock("package",
								tfbuild.AttributeString("name", "openssl"),
							),
						),
					)),
					ExpectError: regexp.MustCompile(`apt not available`),
				},
			},
		})
	})
}

func testAccPackageAptBlock(name string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	return tfbuild.Resource("system_packages_apt", name, attrs...)
}
