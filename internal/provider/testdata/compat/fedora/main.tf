# Reference configuration for the upstream compatibility test (see ../README.md).
# Recorded with a provider binary built from upstream tag v0.5.0 by ../record.sh.
# Do not change without re-recording state.v0.5.0.json for every target.

terraform {
  required_providers {
    system = {
      source = "neuspaces/system"
    }
  }
}

variable "ssh_host" {
  type = string
}

variable "ssh_port" {
  type = number
}

variable "ssh_user" {
  type = string
}

variable "ssh_password" {
  type      = string
  sensitive = true
}

variable "ssh_host_key" {
  type = string
}

provider "system" {
  ssh {
    host     = var.ssh_host
    port     = var.ssh_port
    user     = var.ssh_user
    password = var.ssh_password
    host_key = var.ssh_host_key
  }
}

locals {
  base = "/root/tfps-compat-v050"
}

resource "system_folder" "base" {
  path = local.base
  mode = "755"
}

resource "system_file" "content" {
  path    = "${local.base}/content.txt"
  mode    = "644"
  content = "compat content\n"

  depends_on = [system_folder.base]
}

resource "system_file" "content_sensitive" {
  path              = "${local.base}/sensitive.txt"
  mode              = "600"
  content_sensitive = "compat sensitive content\n"

  depends_on = [system_folder.base]
}

resource "system_file" "source" {
  path   = "${local.base}/source.txt"
  mode   = "644"
  source = "${path.module}/hello.txt"

  depends_on = [system_folder.base]
}

resource "system_link" "link" {
  path   = "${local.base}/link"
  target = system_file.content.path
}

resource "system_group" "group" {
  name = "tfpscompat"
  gid  = 4242
}

resource "system_user" "user" {
  name  = "tfpscompat"
  uid   = 4242
  gid   = system_group.group.gid
  home  = "${local.base}/home"
  shell = "/bin/sh"
}

resource "system_file" "unit" {
  path    = "/etc/systemd/system/tfps-compat.service"
  mode    = "644"
  content = <<-EOT
    [Unit]
    Description=tfps compat busybox web server

    [Service]
    ExecStart=/usr/sbin/busybox httpd -f -p 18080 -h /tmp

    [Install]
    WantedBy=multi-user.target
  EOT
}

resource "system_service_systemd" "service" {
  name    = "tfps-compat"
  enabled = true
  status  = "started"

  depends_on = [system_file.unit]
}

resource "system_systemd_unit" "unit" {
  type   = "service"
  name   = "tfps-compat"
  status = "started"

  depends_on = [system_file.unit, system_service_systemd.service]
}
