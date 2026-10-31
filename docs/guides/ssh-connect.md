---
subcategory: ""
page_title: "SSH connection"
description: |-
    Documentation of SSH connection methods.
---

# SSH connection

The provider attributes presented in the examples below are expected to be combined with provider attributes to configure the [SSH authentication](./ssh-auth).

Since v0.6.0 every connection requires a `host_key`, as shown in the examples below. Refer to [Host key verification](..#host-key-verification) for the alternative `insecure_ignore_host_key`.

## Connection methods

### Direct connection

To connect directly to the remote system, define the connection arguments in the [`ssh` block](..#nestedblock--ssh).

```terraform
provider "system" {
  ssh {
    host = "10.12.13.14"
    port = 22

    # ssh-keyscan -t ed25519 10.12.13.14
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
  }
}
```

### Proxy (bastion host) connection

To connect indirectly to the remote system via a proxy or bastion host, define the connection arguments to the proxy in the [`ssh` block](..#nestedblock--ssh) within the [`proxy` block](..#nestedblock--proxy). Define the connection arguments to the remote system from the perspective of the proxy host in the [`ssh` block](..#nestedblock--ssh).

```terraform
provider "system" {
  proxy {
    ssh {
        host = "10.12.13.14"
        port = 22

        # ssh-keyscan -t ed25519 10.12.13.14
        host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
    }
  }

  ssh {
    host = "192.168.32.4"
    port = 22

    # ssh-keyscan -t ed25519 192.168.32.4
    host_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
  }
}
```
