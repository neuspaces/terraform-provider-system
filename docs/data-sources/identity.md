---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "system_identity | Data Source | terraform-provider-system"
name: "system_identity"
type: "Data Source"
subcategory: ""
description: |-
  system_identity retrieves information about the identity of the user on the remote system.
---

# Data Source: system_identity

`system_identity` retrieves information about the identity of the user on the remote system.

## Usage

```terraform
data "system_identity" "current" {}
```



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `gid` (Number) ID of the primary group of the user.
- `group` (String) Name of the primary group of the user.
- `id` (String) The ID of this resource.
- `uid` (Number) ID of the user.
- `user` (String) Name of the user.

