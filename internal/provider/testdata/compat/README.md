# Upstream compatibility fixtures

These fixtures back `TestAccCompat_upstream_state_empty_plan`
(`internal/provider/compat_test.go`), the drop-in test of this fork: a state file written by the **upstream
`neuspaces/system` v0.5.0 binary** must produce an **empty plan** with this provider after the two migration
operations, which are

1. changing `required_providers.source` from `neuspaces/system` to `dorokhovich-com/system` (the test rewrites it in
   its copy of `main.tf`), and
2. `terraform state replace-provider registry.terraform.io/neuspaces/system <host>/dorokhovich-com/system`.

`plan -detailed-exitcode` has to exit `0`. Exit code `2` (changes) and exit code `1` (error) fail the test.

## Contents

| Path | Purpose |
|---|---|
| `<os>/main.tf` | Reference configuration for the acceptance container of `<os>`, covering every resource type the provider has for that OS |
| `<os>/hello.txt` | Content for `system_file.source`; its etag (`base64(md5(content))`) is what the state records |
| `<os>/state.v0.5.0.json` | The fixture: state of the reference configuration, written by a provider binary built from tag `v0.5.0` |
| `record.sh` | Records `<os>/state.v0.5.0.json` |

The fixture is named `state.v0.5.0.json` and not `terraform.tfstate` on purpose: `.gitignore` ignores `*.tfstate`.

The reference configuration sets every `Optional+Computed` attribute explicitly (`mode`, `uid`, `gid`, `home`,
`shell`) so the fixture does not depend on OS defaults, and it sets the provider attribute `host_key`, so the same
configuration works with v0.5.0 and with this provider, which refuses to connect without a verified host key
(v0.6.0 breaking change).

## Recording

Prerequisites: the acceptance containers of `test/<os>` are running (ports 10022, 11022, 12022), a Go toolchain with
network access for `go mod download`, and a Terraform CLI.

Record with the **oldest supported CLI** (Terraform 1.6), because Terraform refuses to read a state file written by a
newer version, and the test has to run on every supported CLI:

```sh
cd <repository root>
export TERRAFORM=/path/to/terraform-1.6.3
internal/provider/testdata/compat/record.sh alpine 10022 test/alpine/hostkeys/ssh_host_ecdsa_key.pub
internal/provider/testdata/compat/record.sh debian 11022 test/debian/hostkeys/ssh_host_ecdsa_key.pub
internal/provider/testdata/compat/record.sh fedora 12022 test/fedora/hostkeys/ssh_host_ecdsa_key.pub
```

`record.sh` extracts tag `v0.5.0` with `git archive`, builds the provider from it, applies the reference
configuration against the container with `dev_overrides`, copies the resulting state here, and destroys the objects
again. The connection parameters are those of `internal/provider/acctest.yaml` (test containers, test credentials);
`SSH_HOST`, `SSH_USER` and `SSH_PASSWORD` override them.

Verify after recording:

- `"provider": "provider[\"registry.terraform.io/neuspaces/system\"]"` on every resource,
- `"schema_version": 1` on every instance,
- `system_file.source` carries `"source": "etag=…"` and not a local path,
- no real secrets: the only payloads are the test strings `compat content` and `compat sensitive content`, and the
  ssh password is not part of a Terraform state.

## When to re-record

**Re-record when the acceptance container images change** (new base image, new package versions): the fixture
contains values read from the system, such as the installed package version of `grep`, and those legitimately change
with the image.

**Never re-record because the provider code changed.** The whole point of the fixture is that it was written by
upstream `v0.5.0`. If the plan stops being empty after a change to the provider, that is a finding about the
compatibility of the change, not a reason to refresh the fixture: the failure lists the planned changes per resource
address.

In `v0.7.0`, when `source` is updated in place instead of forcing a new resource, exactly one `update` of
`system_file.source` becomes expected (owner decision 1 of the phase 1 plan). That exception is to be added as an
allowlist in `compatPlannedChanges` / the assertion of `compat_test.go` — not by re-recording the fixture.

## Notes

- The test does not run `terraform init`. The provider under test is wired in with `dev_overrides` and is not
  published in any registry, so `init` fails to query `dorokhovich-com/system`; Terraform itself recommends skipping
  `init` under development overrides. The reference configuration has no backend and no modules, so `apply`,
  `state replace-provider` and `plan` work without it.
- For the same reason the test cannot use `plan -out=…`: the CLI refuses to save a plan file while development
  overrides are in effect. The list of planned changes therefore comes from a second `plan -json`.
- `system_systemd_unit` in the reference configuration sets `status`. With neither `status` nor `enabled` set,
  `v0.5.0` builds an empty shell function for the apply command and fails with `systemd unit resource: unexpected
  error` (`internal/client/systemd_unit.go`), so such a configuration cannot be recorded.
