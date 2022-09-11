---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "system_link | Resource | terraform-provider-system"
name: "system_link"
type: "Resource"
subcategory: ""
description: |-
  system_link manages a symbolic link on the remote system.
---

# Resource: system_link

`system_link` manages a symbolic link on the remote system.

## Usage

### Relative target

```terraform
resource "system_link" "relative_target" {
  path   = "/root/link.txt"
  target = "./document.txt"
}
```

### Absolute target

```terraform
resource "system_link" "absolute_target" {
  path   = "/root/link.txt"
  target = "/root/document.txt"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) Path of the link. Must be an absolute path. Not to be confused with the target.
- `target` (String) Target of the link. Can be either an absolute or a relative path. Target is not required to exist when link is created.

### Optional

- `gid` (Number) ID of the group that owns the link. Does *not* change the group owning the target.
- `group` (String) Name of the group that owns the link. Does *not* change the group owning the target.
- `uid` (Number) ID of the user who owns the link. Does *not* change the user owning the target.
- `user` (String) Name of the user who owns the link. Does *not* change the user owning the target.

### Read-Only

- `id` (String) ID of the link

