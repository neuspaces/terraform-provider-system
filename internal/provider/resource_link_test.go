package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"path"
	"strings"
	"sync/atomic"
	"testing"
)

var (
	testLinkId uint32
)

type testLinkConfig struct {
	linkName   string
	targetName string
}

func newTestLinkConfig() testLinkConfig {
	id := atomic.AddUint32(&testLinkId, 1)

	return testLinkConfig{
		linkName:   fmt.Sprintf("link%d", id),
		targetName: fmt.Sprintf("target%d", id),
	}
}

func TestAccLink_create(t *testing.T) {
	testConfig := newTestLinkConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccLinkBlock("test", testRunLinkPath(target, testConfig.linkName), testRunLinkPath(target, testConfig.targetName)),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_link.test", "id", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "path", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "target", testRunLinkPath(target, testConfig.targetName)),
						resource.TestCheckResourceAttr("system_link.test", "user", "root"),
						resource.TestCheckResourceAttr("system_link.test", "uid", "0"),
						resource.TestCheckResourceAttr("system_link.test", "group", "root"),
						resource.TestCheckResourceAttr("system_link.test", "gid", "0"),
					),
				},
			},
		})
	})
}

func TestAccLink_create_user(t *testing.T) {
	t.Skip()
}

func TestAccLink_create_group(t *testing.T) {
	t.Skip()
}

func TestAccLink_update_target(t *testing.T) {
	testConfig := newTestLinkConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccLinkBlock("test", testRunLinkPath(target, testConfig.linkName), testRunLinkPath(target, testConfig.targetName, "a")),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_link.test", "id", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "path", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "target", testRunLinkPath(target, testConfig.targetName, "a")),
					),
				},
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccLinkBlock("test", testRunLinkPath(target, testConfig.linkName), testRunLinkPath(target, testConfig.targetName, "b")),
					)),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("system_link.test", "id", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "path", testRunLinkPath(target, testConfig.linkName)),
						resource.TestCheckResourceAttr("system_link.test", "target", testRunLinkPath(target, testConfig.targetName, "b")),
					),
				},
			},
		})
	})
}

func TestAccLink_update_user(t *testing.T) {
	t.Skip()
}

func TestAccLink_update_group(t *testing.T) {
	t.Skip()
}

func testRunLinkPath(target acctest.Target, nameParts ...string) string {
	return path.Join(target.BasePath, strings.Join(nameParts, ""))
}

func testAccLinkBlock(name string, path string, target string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("path", path),
		tfbuild.AttributeString("target", target),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_link", name, resourceAttrs...)
}
