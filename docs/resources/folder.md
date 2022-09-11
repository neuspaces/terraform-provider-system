---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "system_folder | Resource | terraform-provider-system"
name: "system_folder"
type: "Resource"
subcategory: ""
description: |-
  system_folder manages a folder on the remote system.
---

# Resource: system_folder

`system_folder` manages a folder on the remote system.

`system_folder` ensures a folder exists on the remote server with specific permissions and ownership to user and group.

## Usage

### Minimal

```terraform
resource "system_folder" "minimal" {
  path = "/root/some-folder"
}
```

### Explicit permissions

Folder permissions can be explicitly defined in the `mode` attribute in octal format.

```terraform
resource "system_folder" "permissions" {
  path = "/root/some-folder"
  mode = 700
}
```

### Explicit ownership (using canonical names)

The user and group owning the folder can be configured explicitly in attributes `user` and `group`. The user or group is referenced by their canonical name. The `system_folder` resource expects that the user or group exists on the remote server.

```terraform
resource "system_folder" "ownership_name" {
  path  = "/root/some-folder"
  user  = "johndoe"
  group = "johndoe"
}
```

### Explicit ownership (using uid or gid)

The user and group owning the folder can be configured explicitly in attributes `uid` and `gid`. The user or group is referenced by their id. In contrast to explicit ownership by name, the user or group must not exist on the remote server.

```terraform
resource "system_folder" "ownership_id" {
  path = "/root/some-folder"
  uid  = 1001
  gid  = 1001
}
```

You may consider managing the owning user with a `system_user` resource or the owning group with a `system_group` resource, and use the outputs of these resources to define the `uid` and `gid` attribute.

```terraform
resource "system_group" "johndoe" {
  name = "johndoe"
}

resource "system_user" "johndoe" {
  name = "johndoe"
  gid  = system_group.johndoe.gid
}

resource "system_folder" "ownership_id" {
  path = "/root/some-folder"
  uid  = system_group.johndoe.uid
  gid  = system_group.johndoe.gid
}
```

## Notes

This section describes general notes for using the `system_file` resource.

- The parent folder of the folder references by `path` must exist. It will not be created implicitly by the resource.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) Path to the folder

### Optional

- `gid` (Number) ID of the group that owns the folder
- `group` (String) Name of the group that owns the folder
- `mode` (String) Permissions of the folder in octal format like `755`. Defaults to the umask of the system.
- `uid` (Number) ID of the user who owns the folder
- `user` (String) Name of the user who owns the folder

### Read-Only

- `basename` (String) Base name of the folder. Returns the last element of path. Example: Given the attribute `path` is `/path/to/folder`, the `basename` is `folder`.
- `id` (String) ID of the folder

