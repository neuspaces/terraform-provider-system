# Changelog

All notable changes to this provider are documented in this file. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses semantic versioning.

## [0.6.0] - unreleased

First release of the `dorokhovich-com/system` fork of `neuspaces/system` v0.5.0. Resource, data source
and attribute names are unchanged; see the [migration guide](docs/guides/migration-from-neuspaces.md).

### BREAKING CHANGES

- **provider: connections without `host_key` are refused.** Up to v0.5.0 the provider silently fell back to `ssh.InsecureIgnoreHostKey()` when `host_key` was not set, which left every connection open to man-in-the-middle attacks. Starting with v0.6.0 `ConfigureProvider` fails with `SSH host key verification is not configured` unless either `host_key` is set in the `ssh` block (and in `proxy.ssh` when a proxy is used; `host_key`/`bastion_host_key` in the `connection` block) or the new provider attribute `insecure_ignore_host_key = true` restores the old behavior. With `insecure_ignore_host_key = true` every run prints the warning `SSH host key verification is disabled`. Obtain the host key with `ssh-keyscan -t ed25519 <host>`. See the migration guide (`docs/guides/migration-from-neuspaces.md`).

### Added

- `system_file.overwrite` and `system_folder.overwrite` (Boolean, default `false`): adopt a pre-existing file or folder at `path` on create instead of failing with `file exists`/`folder path exists`. An entry of another kind still fails. Ported from upstream PR #11 by raynigon; fixes upstream issue #4.
- `system_user.create_home` (Boolean, default `false`): create the home directory with the user (`useradd --create-home`). Ported from upstream PR #10 by ii1818nn; fixes upstream issue #9.
- Provider attribute `insecure_ignore_host_key` (Boolean, default `false`), see BREAKING CHANGES.
- Acceptance test `TestAccCompat_upstream_state_empty_plan`: a state file written by upstream v0.5.0 produces an empty plan with this provider after `terraform state replace-provider`.
- CI: GitHub-hosted `ubuntu-latest`, matrix of Terraform 1.6, 1.10, 1.11, 1.14 and OpenTofu 1.6 and current, `govulncheck`, `go vet`, Dependabot for Go modules and GitHub Actions.
- `NOTICE` describing the origin of the fork; this changelog.

### Changed

- `system_file.source` is no longer `ForceNew`: when the referenced content changes (detected via its etag), the remote file is rewritten in place (`update`) instead of being deleted and created again. Configurations that relied on the replacement no longer get it.
- Dependencies: Go 1.27, `terraform-plugin-sdk/v2` 2.40.1, `golang.org/x/crypto` 0.58.0, `golang.org/x/net` 0.62.0, `google.golang.org/grpc` 1.85.0. `govulncheck ./...` reports no vulnerabilities (v0.5.0: 28 reachable, 9 of them in the SSH path).
- Acceptance test images: Debian <version>, Alpine <version>, Fedora <version>.
- Provider address for debugging (`ProviderAddr` in `main.go`) is `registry.terraform.io/dorokhovich-com/system`.

### Fixed

- `system_file`: switching between `content`, `content_sensitive` and `source` in one apply did not write the new content (the first changed attribute, now empty, shadowed the one that carried content).
- Documentation: links from the provider index to the guides pointed to `./docs/guides/...` and were broken in the registry.
- `go vet`: `t.Fatal` called from a non-test goroutine in the acceptance test ssh agent.

### Security

- Host key verification is required (see BREAKING CHANGES).
- All reachable vulnerabilities reported by `govulncheck` on v0.5.0 are fixed by the dependency upgrade.

### Planned for v0.7.0 (announcement)

- `system_file`, `system_folder` and `system_link` move to the Terraform plugin framework behind `terraform-plugin-mux`. Because the framework cannot store `etag=…` in `source`, resources with `source` will show exactly one in-place `update` of `source` on the first plan after upgrading to v0.7.0; no replace, destroy or create. This is the only known deviation from the empty-plan guarantee.
- `system_file.content_wo` (write-only content, Terraform >= 1.11) and `content_wo_version`; `md5sum` is `null` for files managed through `content_wo`.

## [0.5.0] - 2024-05-31

Last upstream release (`neuspaces/system`). See https://github.com/neuspaces/terraform-provider-system/releases/tag/v0.5.0.
