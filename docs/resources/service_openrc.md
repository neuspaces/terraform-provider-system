---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "system_service_openrc | Resource | terraform-provider-system"
name: "system_service_openrc"
type: "Resource"
subcategory: ""
description: |-
  system_service_openrc manages an OpenRC service on the remote system.
---

# Resource: system_service_openrc

`system_service_openrc` manages an OpenRC service on the remote system.

## Usage

### Start a service

This example ensures that the OpenRC service `nginx` is started. The service must exist.

```terraform
resource "system_service_openrc" "nginx" {
  name    = "nginx"
  status  = "started"
}
```

### Enable a service

This example ensures that the OpenRC service `nginx` is enabled. The service is enabled for the runlevel defined in the attribute `runlevel` which defaults to `default`. The service must exist.

```terraform
resource "system_service_openrc" "nginx" {
  name    = "nginx"
  enabled = true
}
```

### Create, enable, and start a service

This example creates an OpenRC service script `busybox-httpd` using the `system_file` resource and subsequently enables and starts the service using the `system_service_openrc`. The service starts a busybox httpd daemon.

```terraform
resource "system_file" "busybox_httpd" {
  path    = "/etc/init.d/busybox-httpd"
  mode    = 755
  user    = "root"
  group   = "root"
  content = <<EOT
#!/sbin/openrc-run

name=$SVCNAME
command="/bin/busybox-extras httpd"
command_args="-p 8080 -h /var/www/html"

depend() {
    need net localmount
    after firewall
}

start_pre() {
    mkdir -p /var/www/html
}
EOT
}

resource "system_service_openrc" "busybox_httpd" {
  name    = system_file.busybox_httpd.basename
  enabled = true
  status  = "started"
}
```

## Notes

This section describes general notes for using the `system_service_openrc` resource.

- The service script must exist.
- The resource does not manage, create, or delete the service script.
- The resource remembers the `enabled` state and the `status` state at the time the resource is created. This state is referred to as the *original state*.
- When the resource is deleted, the service is reverted to the original state.
- Avoid defining multiple `system_service_openrc` resources, which manage the same service in the same Terraform configuration. Instead, merge all attributes in a single `system_service_openrc` resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the service. The service must exist.

### Optional

- `enabled` (Boolean) If `true`, the service will be enabled on the provided runlevel. If not provided, the service will not be changed.
- `reload_on` (Set of String) Set of arbitrary strings which will trigger a reload of the service.
- `restart_on` (Set of String) Set of arbitrary strings which will trigger a restart of the service.
- `runlevel` (String) Runlevel to which the `enabled` attribute refers to. Defaults to `default`.
- `status` (String) Status of the service. If `started`, the service will be started. If `stopped`, the service will be stopped.

### Read-Only

- `id` (String) ID of the service
- `internal` (String, Sensitive)

