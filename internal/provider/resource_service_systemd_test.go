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
	testServiceSystemdId uint32
)

type testServiceSystemdConfig struct {
	serviceName string
	servicePort string
}

func newTestServiceSystemdConfig() testServiceSystemdConfig {
	id := atomic.AddUint32(&testServiceSystemdId, 1)

	return testServiceSystemdConfig{
		serviceName: fmt.Sprintf("httpd-%d", id),
		servicePort: strconv.Itoa(int(8080 + id)),
	}
}

var testAccServiceSystemdOsIds = []string{
	osrelease.DebianId,
	osrelease.FedoraId,
}

// Test to start a service given the service is stopped
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is disabled
//
// Expected:
// - Service is started
// - Service is disabled
func TestAccServiceSystemd_status_started(t *testing.T) {
	testConfig := newTestServiceSystemdConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccServiceSystemdOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceSystemdServiceUnitFileResource(t, target, "test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceSystemdResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_systemd.test"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "status", "started"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_service_systemd.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to stop a service given the service is started
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is started
// - Service is disabled
//
// Expected:
// - Service is stopped
// - Service is disabled
func TestAccServiceSystemd_status_stopped(t *testing.T) {
	testConfig := newTestServiceSystemdConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccServiceSystemdOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceSystemdServiceUnitFileResource(t, target, "test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceSystemdResource("test_prerequisite", testConfig.serviceName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("status"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccServiceSystemdResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_service_systemd", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_systemd.test"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_service_systemd.test", "internal", `{"pre_status":"started","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to enable a service given the service is disabled
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is disabled
//
// Expected:
// - Service is stopped
// - Service is enabled
func TestAccServiceSystemd_enable(t *testing.T) {
	testConfig := newTestServiceSystemdConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccServiceSystemdOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceSystemdServiceUnitFileResource(t, target, "test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceSystemdResource("test", testConfig.serviceName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.AttributeBool("enabled", true),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_systemd.test"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "enabled", "true"),
						provider.TestCheckResourceAttrBase64("system_service_systemd.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to disable a service given the service is enabled
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is enabled
//
// Expected:
// - Service is stopped
// - Service is disabled
func TestAccServiceSystemd_disable(t *testing.T) {
	testConfig := newTestServiceSystemdConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccServiceSystemdOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestServiceSystemdServiceUnitFileResource(t, target, "test", testConfig.serviceName, testConfig.servicePort),
						testAccServiceSystemdResource("test_prerequisite", testConfig.serviceName,
							tfbuild.AttributeBool("enabled", true),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("enabled"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccServiceSystemdResource("test", testConfig.serviceName,
							tfbuild.AttributeBool("enabled", false),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_service_systemd", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_service_systemd.test"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "id", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "name", testConfig.serviceName),
						resource.TestCheckResourceAttr("system_service_systemd.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_service_systemd.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_service_systemd.test", "internal", `{"pre_status":"stopped","pre_enabled":true}`),
					),
				},
			},
		})
	})
}

func testAccTestServiceSystemdServiceUnitFileResource(t *testing.T, target acctest.Target, name string, serviceName string, servicePort string) tfbuild.FileElement {
	var busyboxPath string

	switch target.Os.Id {
	case osrelease.DebianId:
		busyboxPath = "/bin/busybox"
	case osrelease.FedoraId:
		busyboxPath = "/usr/sbin/busybox"
	default:
		t.Skip()
	}

	systemdServiceSpec := heredoc.String(
		fmt.Sprintf(`
			[Unit]
			Description=BusyBox web server
			After=network.target
			
			[Service]
			Type=forking
			User=www-data
			Group=www-data
			ExecStartPre=+/bin/mkdir -p /var/www/html
			ExecStartPre=+/bin/chown www-data:www-data /var/www/html
			ExecStart=%[1]s httpd -p %[2]s -h /var/www/html
			
			[Install]
			WantedBy=multi-user.target
		`,
			busyboxPath,
			servicePort,
		),
	)

	return tfbuild.Resource("system_file", name,
		tfbuild.AttributeString("path", fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)),
		tfbuild.AttributeString("mode", "644"),
		tfbuild.AttributeString("user", "root"),
		tfbuild.AttributeString("group", "root"),
		tfbuild.AttributeString("content", systemdServiceSpec),
	)
}

func testAccServiceSystemdResource(name string, serviceName string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("name", serviceName),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_service_systemd", name, resourceAttrs...)
}
