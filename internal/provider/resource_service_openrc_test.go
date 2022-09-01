package provider_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/heredoc"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"strconv"
	"sync/atomic"
	"testing"
)

var (
	testServiceOpenRcId uint32
)

type testServiceOpenRcConfig struct {
	serviceName string
	servicePort string
}

func newTestServiceOpenRcConfig() testServiceOpenRcConfig {
	id := atomic.AddUint32(&testServiceOpenRcId, 1)

	return testServiceOpenRcConfig{
		serviceName: fmt.Sprintf("httpd-%d", id),
		servicePort: strconv.Itoa(int(8080 + id)),
	}
}

// Test to start a service given the service is stopped
//
// Preconditions:
// - OpenRC service scripts exists at /etc/init.d/httpd-N
// - Service is stopped
// - Service is disabled in runlevel `default`
//
// Expected:
// - Service is started
// - Service is disabled in runlevel `default`
func TestAccServiceOpenRc_status_started(t *testing.T) {
	testConfig := newTestServiceOpenRcConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceOpenrcFileResource("test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceOpenrcResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_openrc.test"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "status", "started"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "enabled", "false"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "runlevel", "default"),
						provider.TestCheckResourceAttrBase64("system_service_openrc.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to stop a service given the service is started
// Preconditions:
// - OpenRC service scripts exists at /etc/init.d/httpd-N
// - Service is started
// - Service is disabled in runlevel `default`
//
// Expected:
// - Service is stopped
// - Service is disabled in runlevel `default`
func TestAccServiceOpenRc_status_stopped(t *testing.T) {
	testConfig := newTestServiceOpenRcConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceOpenrcFileResource("test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceOpenrcResource("test_prerequisite", testConfig.serviceName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("status"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccServiceOpenrcResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_service_openrc", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_openrc.test"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "enabled", "false"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "runlevel", "default"),
						provider.TestCheckResourceAttrBase64("system_service_openrc.test", "internal", `{"pre_status":"started","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to enable a service given the service is disabled
//
// Preconditions:
// - OpenRC service scripts exists at /etc/init.d/httpd-N
// - Service is stopped
// - Service is disabled in runlevel `default`
//
// Expected:
// - Service is stopped
// - Service is enabled in runlevel `default`
func TestAccServiceOpenRc_enable(t *testing.T) {
	testConfig := newTestServiceOpenRcConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceOpenrcFileResource("test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceOpenrcResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.AttributeBool("enabled", true),
							tfbuild.AttributeString("runlevel", "default"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_openrc.test"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "enabled", "true"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "runlevel", "default"),
						provider.TestCheckResourceAttrBase64("system_service_openrc.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to disable a service given the service is enabled
//
// Preconditions:
// - OpenRC service scripts exists at /etc/init.d/httpd-N
// - Service is stopped
// - Service is enabled in runlevel `default`
//
// Expected:
// - Service is stopped
// - Service is disabled in runlevel `default`
func TestAccServiceOpenRc_disable(t *testing.T) {
	testConfig := newTestServiceOpenRcConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, osrelease.AlpineId)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceOpenrcFileResource("test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceOpenrcResource("test_prerequisite", testConfig.serviceName,
							tfbuild.AttributeBool("enabled", true),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("enabled"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccServiceOpenrcResource("test", testConfig.serviceName,
							tfbuild.AttributeBool("enabled", false),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_service_openrc", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_openrc.test"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_openrc.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "enabled", "false"),
						resource.TestCheckResourceAttr("system_service_openrc.test", "runlevel", "default"),
						provider.TestCheckResourceAttrBase64("system_service_openrc.test", "internal", `{"pre_status":"stopped","pre_enabled":true}`),
					),
				},
			},
		})
	})
}

func testAccTestServiceOpenrcFileResource(name string, serviceName string, servicePort string) tfbuild.FileElement {
	openRcServiceSpec := heredoc.String(
		fmt.Sprintf(`
			#!/sbin/openrc-run
			
			name=$SVCNAME
			command="/bin/busybox-extras httpd"
			command_args="-p %[1]s -h /var/www/html"
			
			depend() {
				need net localmount
				after firewall
			}
			
			start_pre() {
				mkdir -p /var/www/html
			}
		`,
			servicePort,
		),
	)

	return tfbuild.Resource("system_file", name,
		tfbuild.AttributeString("path", fmt.Sprintf("/etc/init.d/%s", serviceName)),
		tfbuild.AttributeString("mode", "755"),
		tfbuild.AttributeString("user", "root"),
		tfbuild.AttributeString("group", "root"),
		tfbuild.AttributeString("content", openRcServiceSpec),
	)
}

func testAccServiceOpenrcResource(name string, serviceName string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("name", serviceName),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_service_openrc", name, resourceAttrs...)
}
