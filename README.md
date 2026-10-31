# Terraform Provider for (Operating) System

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/Dorokhovich-COM/terraform-provider-system)](https://github.com/Dorokhovich-COM/terraform-provider-system/releases)
[![main](https://github.com/Dorokhovich-COM/terraform-provider-system/actions/workflows/main.yml/badge.svg)](https://github.com/Dorokhovich-COM/terraform-provider-system/actions/workflows/main.yml)
[![GitHub License](https://img.shields.io/github/license/Dorokhovich-COM/terraform-provider-system)](https://github.com/Dorokhovich-COM/terraform-provider-system/blob/main/LICENSE)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-5c4ee5.svg)](https://registry.terraform.io/providers/dorokhovich-com/system/latest)

Releases: [registry.terraform.io](https://registry.terraform.io/providers/dorokhovich-com/system/latest)

Documentation: [registry.terraform.io](https://registry.terraform.io/providers/dorokhovich-com/system/latest/docs)

Changelog: [CHANGELOG.md](CHANGELOG.md)

The Terraform Provider for (Linux Operating) System allows managing files, directories, users, groups, packages, and services on remote servers on operating system level agent-less via SSH.

> Even in a cloud-native heaven ☁️, there will still be use cases for pets 🐈

## Origin

This provider is a maintained fork of [neuspaces/terraform-provider-system](https://github.com/neuspaces/terraform-provider-system) (v0.5.0, MPL-2.0); see [NOTICE](NOTICE) and the [migration guide](docs/guides/migration-from-neuspaces.md).

## Highlights

* Manage files, directories, users, groups, packages, and services on remote servers
* Connect to and authenticate with remote servers via SSH
* No agent on remote server required
* Seamless integration with Terraform providers of [all major IaaS cloud providers](examples)
* Support for Debian, Alpine, and Fedora Linux confirmed via acceptance test suite

## Quick Starts

- [Using the provider](https://registry.terraform.io/providers/dorokhovich-com/system/latest/docs)
- [Examples](examples)

## Use Cases

The provider aims to allow configuring individual remote servers according to *mutable infrastructure* approach. You might find your use case in the following non-exhaustive list:

* Individual servers or virtual machines which are not recreated when configuration changes
* Share or distribute server or virtual machine configuration as using Terraform modules

The provider is not suitable for *immutable infrastructure* approaches such as fleets of homogeneous virtual machines. In this case, you may consider a more suitable configuration mechanism.

## User documentation

Refer to the comprehensive [user documentation of the provider in the Terraform Registry](https://registry.terraform.io/providers/dorokhovich-com/system/latest/docs).

## Frequently Asked Questions

Responses to the most frequently asked questions can be found in the [FAQ](https://registry.terraform.io/providers/dorokhovich-com/system/latest/docs/guides/faq).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.6+ or [OpenTofu](https://opentofu.org/docs/intro/install/) 1.6+
