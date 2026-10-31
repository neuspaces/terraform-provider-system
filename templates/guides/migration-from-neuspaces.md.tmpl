---
subcategory: ""
page_title: "Migrating from neuspaces/system"
description: |-
    How to move an existing configuration and state from the neuspaces/system provider to dorokhovich-com/system.
---

# Migrating from `neuspaces/system`

`dorokhovich-com/system` is a fork of `neuspaces/system` v0.5.0 with the full upstream history. Every resource,
data source and attribute of v0.5.0 exists with the same name, type and import format, and every resource keeps
schema version 1, so an existing state file is read as is.

## Step 1: point the configuration at the new provider

```terraform
terraform {
  required_providers {
    system = {
      source  = "dorokhovich-com/system"
      version = ">= 0.6.0"
    }
  }
}
```

## Step 2: move the state

```shell
terraform init -upgrade
terraform state replace-provider registry.terraform.io/neuspaces/system registry.terraform.io/dorokhovich-com/system
terraform plan
```

OpenTofu users replace `registry.terraform.io` with `registry.opentofu.org` in both arguments.

## Step 3 (only if your configuration has no `host_key`): verify the host key

Since v0.6.0 the provider refuses to connect to an ssh server whose host key is not configured and fails with
`SSH host key verification is not configured`. Releases up to v0.5.0 silently ignored the host key, which exposed
every connection to man-in-the-middle attacks. Choose one of two exits:

1. Recommended: set `host_key` in the `ssh` block (and in `proxy.ssh` when a proxy is used; `host_key` and
   `bastion_host_key` in the `connection` block). Obtain the key with `ssh-keyscan -t ed25519 <host>`.

   ```terraform
   provider "system" {
     ssh {
       host     = "10.12.13.14"
       user     = "root"
       host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
     }
   }
   ```

2. Keep the old behavior: set `insecure_ignore_host_key = true` on the provider. The provider then warns
   `SSH host key verification is disabled` on every run. The host key is not verified, so every connection
   stays open to a man-in-the-middle, and the provider runs arbitrary commands as root over that connection.

   ```terraform
   provider "system" {
     insecure_ignore_host_key = true

     ssh {
       host = "10.12.13.14"
       user = "root"
     }
   }
   ```

This is the only configuration change the migration may require and the only known deviation from a drop-in
replacement in v0.6.0.

## What the plan shows after the migration

With `host_key` or `insecure_ignore_host_key` configured, `terraform plan` is empty for every resource of v0.5.0.
The new attributes `system_file.overwrite`, `system_folder.overwrite` and `system_user.create_home` default to
`false`, which reproduces the v0.5.0 behavior, and do not appear as changes.

## Behavior changes to know about

- `system_file.source`: a change of the referenced content updates the file in place. v0.5.0 deleted and
  re-created the file.
- Looking ahead to v0.7.0: `system_file`, `system_folder` and `system_link` move to the plugin framework. The
  first plan after that upgrade shows exactly one in-place `update` of `source` for each `system_file` that uses
  `source` (the state value `etag=…` is normalized); no resource is replaced.
