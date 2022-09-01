---
subcategory: ""
page_title: "Frequently Asked Questions"
description: |-
    
---

# FAQ

This page answers frequently asked questions. If questions you encountered remain open, feel invited to ask them in the [Q&A forum](https://github.com/neuspaces/terraform-provider-system/discussions/categories/q-a).

## Why is the provider called *system*?

The provider aims to apply configuration on an operating *system* level. Alternative names like *operating-system* or *os* have been considered.

## Why does this provider exist in a world of cloud-native services and immutable infrastructure?

Cloud-native services certainly cover many use cases and provide benefits to their owners over traditional mutable infrastructure. Immutable infrastructure is the preferred approach for provisioning and running production-grade systems at scale.

The existence of this provider is justified for use cases in which cloud-native or immutable infrastructure cannot be applied for practical or economic reasons.

For examples, refer to the question on use cases for this provider.

## What are the use cases of this provider?

You may use this provider to manage operating system related resources:

- on individual heterogeneous servers or virtual machines, or
- on servers which frequently change but cannot be destroyed and recreated easily, or
- to prepare images of virtual machines as part of an immutable infrastructure approach

You may also use this provider as a replacement for the [Terraform built-in ssh provisioner](https://www.terraform.io/language/resources/provisioners/connection) as far as the available resources of this provider cover your use case.

Specific examples for use cases are:
- Individual on-premises servers or virtual machines
- Prototyping with manage IaaS virtual machines
- Bootstraping a small [k3s](https://k3s.io/) based Kubernetes cluster
- Configuration of embedded systems such as Raspberry Pi etc.

In summary, the provider is meant to take care of your pets and not of your cattle.

### For which use cases should this provider be avoided?

The provider is not intended for use cases which cover:

- serverless and cloud-native infrastructure
- immutable infrastructure
- managing fleets of homogeneous virtual machines

## Which operating systems are supported?

Refer to the page on [supported systems](./supported-systems).

## How do I report a bug or submit a feature request?

Bugs and feature requests are managed in [GitHub issues](https://github.com/neuspaces/terraform-provider-system/issues).

If you encountered a bug or if you have a feature in mind, search in the [open issues on GitHub](https://github.com/neuspaces/terraform-provider-system/issues) for similar reports or requests.

If you discovered a new bug, open a [bug issue on GitHub](https://github.com/neuspaces/terraform-provider-system/issues/new?assignees=&labels=bug&template=bug.md).

If you want to share a feature request, open an [enhancement issue on GitHub](https://github.com/neuspaces/terraform-provider-system/issues/new?assignees=&labels=bug&template=enhancement.md).
