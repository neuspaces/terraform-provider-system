---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "system_file | Data Source | terraform-provider-system"
name: "system_file"
type: "Data Source"
subcategory: ""
description: |-
  system_file retrieves meta information about and content of a file on the remote system.
---

# Data Source: system_file

`system_file` retrieves meta information about and content of a file on the remote system.

~> `system_file` always reads and stores the content of the file in the state. Therefore, `system_file` is not recommended for large files or files which contain sensitive content.

-> Use `system_file_meta` if you only need meta information about the file.

## Usage

```terraform
data "system_file" "hostname" {
    path = "/etc/hostname"
}
```



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) Absolute path to the file.

### Read-Only

- `basename` (String) Base name of the file. Returns the last element of path. Example: Given the attribute `path` is `/path/to/file.txt`, the `basename` is `file.txt`.
- `content` (String, Sensitive) Content of the file
- `gid` (Number) ID of the group that owns the file
- `group` (String) Name of the group that owns the file
- `id` (String) ID of the file
- `md5sum` (String) MD5 checksum of the remote file contents on the system in base64 encoding.
- `mode` (String) Permissions of the file in octal format like `755`.
- `uid` (Number) ID of the user who owns the file
- `user` (String) Name of the user who owns the file

