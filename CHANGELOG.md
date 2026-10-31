# Changelog

All notable changes to this provider are documented in this file. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses semantic versioning.

## [0.6.0] - unreleased

First release of the `dorokhovich-com/system` fork of `neuspaces/system` v0.5.0. Resource, data source
and attribute names are unchanged; see the [migration guide](docs/guides/migration-from-neuspaces.md).

### BREAKING CHANGES

- **provider: connections without `host_key` are refused.** Up to v0.5.0 the provider silently fell back to `ssh.InsecureIgnoreHostKey()` when `host_key` was not set, which left every connection open to man-in-the-middle attacks. Starting with v0.6.0 `ConfigureProvider` fails with `SSH host key verification is not configured` unless either `host_key` is set in the `ssh` block (and in `proxy.ssh` when a proxy is used; `host_key`/`bastion_host_key` in the `connection` block) or the new provider attribute `insecure_ignore_host_key = true` restores the old behavior. With `insecure_ignore_host_key = true` every run prints the warning `SSH host key verification is disabled`, naming the block whose key is missing. Obtain the host key with `ssh-keyscan -t ed25519 <host>`. See the migration guide (`docs/guides/migration-from-neuspaces.md`).

### Added

- `system_file.overwrite` and `system_folder.overwrite` (Boolean, default `false`): adopt a pre-existing file or folder at `path` on create instead of failing with `file exists`/`folder path exists`. An entry of another kind still fails, including a symbolic link, even one pointing at an entry of the right kind: nothing is ever written or changed through a link on another path. Ported from upstream PR #11 by raynigon; fixes upstream issue #4.
- `system_user.create_home` (Boolean, default `false`): create the home directory with the user (`useradd --create-home`). Ported from upstream PR #10 by ii1818nn; fixes upstream issue #9. With `create_home = false` (the default) the provider keeps passing `--no-create-home`, so the default does not depend on `CREATE_HOME` in `/etc/login.defs` (on Fedora it is `yes`); this preserves the v0.5.0 behavior.
- Provider attribute `insecure_ignore_host_key` (Boolean, default `false`, environment variable `TF_PROVIDER_SYSTEM_INSECURE_IGNORE_HOST_KEY`), see BREAKING CHANGES. When the value comes from the environment variable rather than from the configuration, the warning says so, because an environment variable leaves no trace in the configuration under review.
- Acceptance test `TestAccCompat_upstream_state_empty_plan`: a state file written by a provider binary built from upstream tag v0.5.0 produces an empty plan with this provider after `terraform state replace-provider`. It runs on all three target systems against the whole CLI matrix.
- CI: GitHub-hosted `ubuntu-latest`, matrix of Terraform 1.6, 1.10, 1.11, 1.14 and 1.17 and OpenTofu 1.6 and current, `govulncheck`, `go vet`, `tfplugindocs validate`, a generated-files check, and Dependabot for Go modules and GitHub Actions.
- `NOTICE` describing the origin of the fork; this changelog.

### Changed

- `system_file.source` is no longer `ForceNew`: when the referenced content changes (detected via its etag), the remote file is rewritten in place (`update`) instead of being deleted and created again. Configurations that relied on the replacement no longer get it.
- `system_file`: removing `source` from an existing resource no longer replaces it with an empty file. Up to v0.5.0 this recreated the file with `touch` and discarded its content; it is now an in-place update which keeps the remote content and manages only permissions and ownership.
- Dependencies: Go 1.27, `terraform-plugin-sdk/v2` 2.40.1, `golang.org/x/crypto` 0.58.0, `golang.org/x/net` 0.62.0, `google.golang.org/grpc` 1.85.0. `govulncheck ./...` reports no reachable vulnerabilities (v0.5.0: 28 reachable, 9 of them in the SSH path); one advisory remains in an imported package and one in a required module that the provider does not call.
- Acceptance test images: Debian <version>, Alpine <version>, Fedora <version>.
- Supported Terraform versions: 1.6 and later, and OpenTofu 1.6 and later (the README previously stated Terraform 0.12+). The plugin protocol stays 5; existing state files are unaffected.
- Provider address for debugging (`ProviderAddr` in `main.go`) is `registry.terraform.io/dorokhovich-com/system`.
- Release pipeline: the GoReleaser configuration and action are migrated to GoReleaser v2 and pinned to v2.19.0. The configuration was v1 syntax, so the release would have failed on the first tag. The artifact set (per-platform zip, `SHA256SUMS` with a detached GPG signature, registry manifest with protocol 5) is unchanged apart from the next entry.
- Release artifacts: the `windows/arm` (32-bit ARM) archive is no longer published, because the Go toolchain this release is built with has no `windows/arm` port. `windows/386`, `windows/amd64` and `windows/arm64` are unchanged.
- Acceptance test container images are built inside the workflow instead of being pulled from `ghcr.io/neuspaces/terraform-provider-system/*-acctest`, which this repository cannot read. The `build-{alpine,debian,fedora}-test-container.yml` workflows publish them to `ghcr.io/dorokhovich-com/terraform-provider-system/<os>-acctest` and can be started manually.

### Fixed

- `system_file`: switching between `content`, `content_sensitive` and `source` in one apply did not write the new content (the first changed attribute, now empty, shadowed the one that carried content).
- `system_file` and `system_folder`: imported resources now carry the default value of `overwrite`, so a configuration which omits the attribute produces an empty plan right after `terraform import`.
- `system_folder`: a folder adopted with `overwrite = true` now gets the configured `mode`. `mkdir -m` does not change the permissions of a directory which already exists.
- `system_file`: updates triggered by a change of `source` are gzip-compressed on the wire, like the initial upload.
- Documentation: links from the provider index to the guides pointed to `./docs/guides/...` and were broken in the registry; the index frontmatter no longer carries the `layout` key, which `tfplugindocs validate` rejects.
- Documentation: the FAQ and examples guides and the question issue template pointed at the upstream repository, including its discussions, which are disabled here. The FAQ link for feature requests also carried the `bug` label.
- `go vet`: `t.Fatal` called from a non-test goroutine in the acceptance test ssh agent, which also made a failure there silent instead of failing the test.

### Security

- Host key verification is required (see BREAKING CHANGES).
- All reachable vulnerabilities reported by `govulncheck` on v0.5.0 are fixed by the dependency upgrade. `govulncheck` runs in CI on every push and pull request and fails the build on vulnerabilities reachable from the provider code; vulnerabilities in imported packages and required modules which the provider does not call are printed with their identifiers and do not fail the build.
- Dependabot does not propose updates for `gopkg.in/yaml.v3`. The module is held before v3.0.1 by a `replace` directive because v3.0.1 merges YAML merge keys shallowly, which silently empties the acceptance test target configuration.

### Planned for v0.7.0 (announcement)

- `system_file`, `system_folder` and `system_link` move to the Terraform plugin framework behind `terraform-plugin-mux`. Because the framework cannot store `etag=…` in `source`, resources with `source` will show exactly one in-place `update` of `source` on the first plan after upgrading to v0.7.0; no replace, destroy or create. This is the only known deviation from the empty-plan guarantee (the v0.6.0 `host_key` requirement is a configuration change, not a plan change).
- `system_file.content_wo` (write-only content, Terraform >= 1.11) and `content_wo_version`; `md5sum` is `null` for files managed through `content_wo`.

## [0.5.0] - 2024-05-31

Last upstream release (`neuspaces/system`). See https://github.com/neuspaces/terraform-provider-system/releases/tag/v0.5.0.
