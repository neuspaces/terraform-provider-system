package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"strings"
	"sync/atomic"
	"testing"
)

var (
	testUserId uint32
)

type testUserConfig struct {
	userName string
}

func newTestUserConfig() testUserConfig {
	id := atomic.AddUint32(&testUserId, 1)

	return testUserConfig{
		userName: fmt.Sprintf("user%d", id),
	}
}

func TestAccUser_create(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "uid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "system", "false"),
						resource.TestCheckResourceAttrSet("system_user.test", "home"),
						resource.TestCheckResourceAttrSet("system_user.test", "shell"),
					),
				},
			},
		})
	})
}

func TestAccUser_update_name(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "a")),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "b"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "b")),
					),
				},
			},
		})
	})
}

func TestAccUser_update_group(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test_a", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test_a", "name")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "a")),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test_a", testRunGroupName(testConfig.userName, "a")),
						testAccGroupBlock("test_b", testRunGroupName(testConfig.userName, "b")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "b"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test_b", "name")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "b")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "b")),
					),
				},
			},
		})
	})
}

func TestAccUser_update_gid(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test_a", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"),
							tfbuild.AttributeTraversal("gid", tfbuild.TraversalResourceAttribute("system_group", "test_a", "id")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "a")),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test_a", testRunGroupName(testConfig.userName, "a")),
						testAccGroupBlock("test_b", testRunGroupName(testConfig.userName, "b")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "b"),
							tfbuild.AttributeTraversal("gid", tfbuild.TraversalResourceAttribute("system_group", "test_b", "id")),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "b")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "b")),
					),
				},
			},
		})
	})
}

func TestAccUser_system_create(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a"),
							tfbuild.AttributeBool("system", true),
						),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"),
							tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
							tfbuild.AttributeBool("system", true),
						),
					))),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_user.test"),
						resource.TestCheckResourceAttrSet("system_user.test", "id"),
						resource.TestCheckResourceAttr("system_user.test", "name", testRunUserName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "uid"),
						resource.TestCheckResourceAttr("system_user.test", "group", testRunGroupName(testConfig.userName, "a")),
						resource.TestCheckResourceAttrSet("system_user.test", "gid"),
						resource.TestCheckResourceAttr("system_user.test", "system", "true"),
					),
				},
			},
		})
	})
}

func testRunUserName(userName string, extensions ...string) string {
	return "test" + acctest.Current().Id + userName + strings.Join(extensions, "")
}

func testAccUserBlock(resourceName string, name string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("name", name),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_user", resourceName, resourceAttrs...)
}
