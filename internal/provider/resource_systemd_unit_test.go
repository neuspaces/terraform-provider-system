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
	testSystemdUnitId uint32
)

type testSystemdUnitConfig struct {
	unitName        string
	unitServicePort string
}

func newTestSystemdUnitConfig() testSystemdUnitConfig {
	id := atomic.AddUint32(&testSystemdUnitId, 1)

	return testSystemdUnitConfig{
		unitName:        fmt.Sprintf("unit-httpd-%d", id),
		unitServicePort: strconv.Itoa(int(8080 + id)),
	}
}

var testAccSystemdUnitOsIds = []string{
	osrelease.DebianId,
	osrelease.FedoraId,
}

// Test to start a service unit given the unit is stopped
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is disabled
//
// Expected:
// - Service is started
// - Service is disabled
func TestAccSystemdUnit_status_started(t *testing.T) {
	testConfig := newTestSystemdUnitConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccSystemdUnitOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestSystemdUnitServiceUnitFileResource(t, target, "test", testConfig.unitName, testConfig.unitServicePort),
						testAccSystemdUnitResource("test", testConfig.unitName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_systemd_unit.test"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "id", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "name", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "status", "started"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_systemd_unit.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to stop a service unit given the unit is started
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is started
// - Service is disabled
//
// Expected:
// - Service is stopped
// - Service is disabled
func TestAccSystemdUnit_status_stopped(t *testing.T) {
	testConfig := newTestSystemdUnitConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccSystemdUnitOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestSystemdUnitServiceUnitFileResource(t, target, "test", testConfig.unitName, testConfig.unitServicePort),
						testAccSystemdUnitResource("test_prerequisite", testConfig.unitName,
							tfbuild.AttributeString("status", "started"),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("status"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccSystemdUnitResource("test", testConfig.unitName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_systemd_unit", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_systemd_unit.test"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "id", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "name", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_systemd_unit.test", "internal", `{"pre_status":"started","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to enable a service unit given the unit is disabled
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is disabled
//
// Expected:
// - Service is stopped
// - Service is enabled
func TestAccSystemdUnit_enable(t *testing.T) {
	testConfig := newTestSystemdUnitConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccSystemdUnitOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestSystemdUnitServiceUnitFileResource(t, target, "test", testConfig.unitName, testConfig.unitServicePort),
						testAccSystemdUnitResource("test", testConfig.unitName,
							tfbuild.AttributeString("status", "stopped"),
							tfbuild.AttributeBool("enabled", true),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_systemd_unit.test"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "id", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "name", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "enabled", "true"),
						provider.TestCheckResourceAttrBase64("system_systemd_unit.test", "internal", `{"pre_status":"stopped","pre_enabled":false}`),
					),
				},
			},
		})
	})
}

// Test to disable a service unit given the unit is enabled
//
// Preconditions:
// - Systemd unit file exists at /etc/systemd/system/httpd-N
// - Service is stopped
// - Service is enabled
//
// Expected:
// - Service is stopped
// - Service is disabled
func TestAccSystemdUnit_disable(t *testing.T) {
	testConfig := newTestSystemdUnitConfig()

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		t.Parallel()

		acctest.SkipWhenOSNotEquals(t, target, testAccSystemdUnitOsIds...)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctest.ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: tfbuild.FileString(tfbuild.File(
						acctest.ProviderConfigBlock(target.Configs.Default()),
						testAccTestSystemdUnitServiceUnitFileResource(t, target, "test", testConfig.unitName, testConfig.unitServicePort),
						testAccSystemdUnitResource("test_prerequisite", testConfig.unitName,
							tfbuild.AttributeBool("enabled", true),
							tfbuild.InnerBlock("lifecycle",
								tfbuild.Attribute("ignore_changes", tfbuild.List(tfbuild.Identifier("enabled"))),
							),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_file", "test"),
							),
						),
						testAccSystemdUnitResource("test", testConfig.unitName,
							tfbuild.AttributeBool("enabled", false),
							tfbuild.DependsOn(
								tfbuild.TraversalResource("system_systemd_unit", "test_prerequisite"),
							),
						),
					)),
					Check: resource.ComposeTestCheckFunc(
						provider.TestLogResourceAttr(t, "system_systemd_unit.test"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "id", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "name", testConfig.unitName),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "status", "stopped"),
						resource.TestCheckResourceAttr("system_systemd_unit.test", "enabled", "false"),
						provider.TestCheckResourceAttrBase64("system_systemd_unit.test", "internal", `{"pre_status":"stopped","pre_enabled":true}`),
					),
				},
			},
		})
	})
}

func testAccTestSystemdUnitServiceUnitFileResource(t *testing.T, target acctest.Target, name string, unitName string, unitServicePort string) tfbuild.FileElement {
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
			unitServicePort,
		),
	)

	return tfbuild.Resource("system_file", name,
		tfbuild.AttributeString("path", fmt.Sprintf("/etc/systemd/system/%s.service", unitName)),
		tfbuild.AttributeString("mode", "644"),
		tfbuild.AttributeString("user", "root"),
		tfbuild.AttributeString("group", "root"),
		tfbuild.AttributeString("content", systemdServiceSpec),
	)
}

func testAccSystemdUnitResource(name string, unitName string, attrs ...tfbuild.BlockElement) tfbuild.FileElement {
	resourceAttrs := []tfbuild.BlockElement{
		tfbuild.AttributeString("type", "service"),
		tfbuild.AttributeString("name", unitName),
	}
	resourceAttrs = append(resourceAttrs, attrs...)

	return tfbuild.Resource("system_systemd_unit", name, resourceAttrs...)
}
