package provider_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"path"
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

// testAccCheckRemoteFolder asserts over ssh whether a folder exists on the target
func testAccCheckRemoteFolder(target acctest.Target, folderPath string, expectExists bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		_, err := client.NewFolderClient(target.Provider.System).Get(context.Background(), folderPath)
		switch {
		case expectExists && err != nil:
			return fmt.Errorf("expected folder %s to exist: %v", folderPath, err)
		case !expectExists && err == nil:
			return fmt.Errorf("expected folder %s to be absent", folderPath)
		case !expectExists && !errors.Is(err, client.ErrFolderNotFound):
			return err
		}
		return nil
	}
}

func testAccUserCreateHomeCase(t *testing.T, createHome *bool, system bool, expectHome bool) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		homePath := path.Join(target.BasePath, "home-"+testConfig.userName)

		userAttrs := []tfbuild.BlockElement{
			tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
			tfbuild.AttributeString("home", homePath),
		}
		groupAttrs := []tfbuild.BlockElement{}
		if system {
			userAttrs = append(userAttrs, tfbuild.AttributeBool("system", true))
			groupAttrs = append(groupAttrs, tfbuild.AttributeBool("system", true))
		}
		expectedAttr := "false"
		if createHome != nil {
			userAttrs = append(userAttrs, tfbuild.AttributeBool("create_home", *createHome))
			if *createHome {
				expectedAttr = "true"
			}
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a"), groupAttrs...),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"), userAttrs...),
					))),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_user.test", "home", homePath),
						resource.TestCheckResourceAttr("system_user.test", "create_home", expectedAttr),
						testAccCheckRemoteFolder(target, homePath, expectHome),
					),
				},
			},
		})
	})
}

func TestAccUser_create_home(t *testing.T) {
	v := true
	testAccUserCreateHomeCase(t, &v, false, true)
}

func TestAccUser_create_home_false(t *testing.T) {
	v := false
	testAccUserCreateHomeCase(t, &v, false, false)
}

func TestAccUser_create_home_default_false(t *testing.T) {
	testAccUserCreateHomeCase(t, nil, false, false)
}

func TestAccUser_create_home_system(t *testing.T) {
	v := true
	testAccUserCreateHomeCase(t, &v, true, true)
}

// testAccUserAttr reads an attribute of a user resource from the state into out
func testAccUserAttr(resourceName string, attr string, out *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		v, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %s of %s not found in state", attr, resourceName)
		}
		*out = v
		return nil
	}
}

// testAccUserAttrUnchanged asserts that an attribute of a user resource still has the value in want
func testAccUserAttrUnchanged(resourceName string, attr string, want *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var got string
		if err := testAccUserAttr(resourceName, attr, &got)(s); err != nil {
			return err
		}
		if *want == "" {
			return fmt.Errorf("no previous value of %s of %s recorded", attr, resourceName)
		}
		if got != *want {
			return fmt.Errorf("expected %s of %s to be unchanged, but it changed from %s to %s; the user has been recreated", attr, resourceName, *want, got)
		}
		return nil
	}
}

// TestAccUser_create_home_add_to_existing asserts that adding create_home to an existing user
// neither recreates the user nor creates the home directory after the fact.
//
// The attribute is deliberately not ForceNew: recreating the user would run userdel, which also
// removes the primary group of the user.
func TestAccUser_create_home_add_to_existing(t *testing.T) {
	testConfig := newTestUserConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()
		homePath := path.Join(target.BasePath, "home-"+testConfig.userName)

		userAttrs := []tfbuild.BlockElement{
			tfbuild.AttributeTraversal("group", tfbuild.TraversalResourceAttribute("system_group", "test", "name")),
			tfbuild.AttributeString("home", homePath),
		}
		userAttrsCreateHome := append(append([]tfbuild.BlockElement{}, userAttrs...), tfbuild.AttributeBool("create_home", true))

		// uid of the user after the first step
		var uid string

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"), userAttrs...),
					))),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_user.test", "create_home", "false"),
						resource.TestCheckResourceAttrSet("system_user.test", "uid"),
						testAccUserAttr("system_user.test", "uid", &uid),
						testAccCheckRemoteFolder(target, homePath, false),
					),
				},
				{
					Config: provider.TestLogString(t, tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.userName, "a")),
						testAccUserBlock("test", testRunUserName(testConfig.userName, "a"), userAttrsCreateHome...),
					))),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_user.test", "create_home", "true"),
						// the user is updated in place, not recreated
						testAccUserAttrUnchanged("system_user.test", "uid", &uid),
						// create_home only applies when the user is created
						testAccCheckRemoteFolder(target, homePath, false),
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
