package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"strings"
	"sync/atomic"
	"testing"
)

var (
	testGroupId uint32
)

type testGroupConfig struct {
	groupName string
}

func newTestGroupConfig() testGroupConfig {
	id := atomic.AddUint32(&testGroupId, 1)

	return testGroupConfig{
		groupName: fmt.Sprintf("group%d", id),
	}
}

func TestAccGroup_create(t *testing.T) {
	testConfig := newTestGroupConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.groupName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("system_group.test", "id"),
						resource.TestCheckResourceAttr("system_group.test", "name", testRunGroupName(testConfig.groupName)),
						resource.TestCheckResourceAttrSet("system_group.test", "gid"),
						resource.TestCheckResourceAttr("system_group.test", "system", "false"),
					),
				},
			},
		})
	})
}

func TestAccGroup_update_name(t *testing.T) {
	testConfig := newTestGroupConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.groupName, "a")),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("system_group.test", "id"),
						resource.TestCheckResourceAttr("system_group.test", "name", testRunGroupName(testConfig.groupName, "a")),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.groupName, "b")),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("system_group.test", "id"),
						resource.TestCheckResourceAttr("system_group.test", "name", testRunGroupName(testConfig.groupName, "b")),
					),
				},
			},
		})
	})
}

func TestAccGroup_system_create(t *testing.T) {
	testConfig := newTestGroupConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccGroupBlock("test", testRunGroupName(testConfig.groupName),
							tfbuild.AttributeBool("system", true),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("system_group.test", "id"),
						resource.TestCheckResourceAttr("system_group.test", "name", testRunGroupName(testConfig.groupName)),
						resource.TestCheckResourceAttrSet("system_group.test", "gid"),
						resource.TestCheckResourceAttr("system_group.test", "system", "true"),
					),
				},
			},
		})
	})
}

func testRunGroupName(groupName string, extensions ...string) string {
	return "test" + acctest.Current().Id + groupName + strings.Join(extensions, "")
}

func testAccGroupBlock(resourceName string, name string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("name", name),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_group", resourceName, resourceAttrs...)
}
