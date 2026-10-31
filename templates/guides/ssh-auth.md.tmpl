---
subcategory: ""
page_title: "SSH authentication"
description: |-
    Documentation of SSH authentication methods.
---

# SSH authentication

The provider supports the SSH authentication methods which are documented on this page.

The provider attributes presented in the examples below are expected to be combined with provider attributes to configure the [SSH connection](./ssh-connect). Since v0.6.0 every connection requires a `host_key`; refer to [Host key verification](..#host-key-verification).

!> Hard-coded credentials are not recommended in any Terraform configuration and risks secret leakage should this file ever be committed to a public version control system.

## Authentication methods

### Agent

```terraform
provider "system" {
  ssh {
    user  = "root"
    agent = true

    # ssh-keyscan -t ed25519 10.12.13.14
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
  }
}
```

### Password

```terraform
provider "system" {
  ssh {
    user     = "root"
    password = "s3cr3t"

    # ssh-keyscan -t ed25519 10.12.13.14
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
  }
}
```

### Private key

```terraform
provider "system" {
  ssh {
    user        = "root"
    private_key = file("./root.key")

    # ssh-keyscan -t ed25519 10.12.13.14
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
  }
}
```

## Privilege escalation (sudo)

The provider supports privilege escalation on the remote system via sudo. Enable `sudo` to connect to the remote system with an unprivileged used and execute commands as root.

If provider attribute `sudo` is `true`, commands are executed on the remote system using `sudo`. As a prerequisite sudo must be installed and configured on the remote system. The user must be able to run sudo without password prompt (NOPASSWD).

```terraform
provider "system" {
  ssh {
    user        = "user"
    private_key = file("./user.key")

    # ssh-keyscan -t ed25519 10.12.13.14
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."

    sudo = true
  }
}
```