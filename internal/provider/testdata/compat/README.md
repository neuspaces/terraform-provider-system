# Released-version compatibility fixtures

These fixtures back `TestAccCompat_v050_state_empty_plan` (`internal/provider/compat_test.go`): a state file written
by the **released `v0.5.0` binary** must produce an **empty plan** with the provider under test, for the same
configuration, without any change to `required_providers`.

`plan -detailed-exitcode` has to exit `0`. Exit code `2` (changes) and exit code `1` (error) fail the test.

The fixtures are recorded with Terraform, so the provider address in their state is
`registry.terraform.io/neuspaces/system`. Under OpenTofu the test moves the state to the OpenTofu registry host with
`state replace-provider` before planning; the namespace and type stay the same, so this is a CLI difference and not
a change of provider.

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

Record with the **oldest supported CLI** (Terraform 1.6), so that the fixture cannot depend on anything a newer CLI
writes into a state file. A newer CLI reads it without complaint, and the test runs on every supported CLI:

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

Re-recording is in scope only when the **acceptance container images** change, and even then it is needed less often
than one would expect: values the provider reads from the system, such as the installed version of `grep`, are
computed attributes that the refresh of the plan simply updates, so a new package version does not make the plan
non-empty. What does break the fixture is a change that requires the reference configuration to be edited, for
example a different path to `busybox` in the unit file.

**Never re-record because the provider code changed.** The whole point of the fixture is that it was written by
upstream `v0.5.0`. If the plan stops being empty after a change to the provider, that is a finding about the
compatibility of the change, not a reason to refresh the fixture: the failure lists the changed resources per
resource address.

Because the two cases look the same from the test output, decide between them with the upstream binary instead of by
reasoning. Against the new image, run a plan with the **old fixture** and the **`v0.5.0` binary** (`record.sh` builds
that binary from the tag; the first half of the script does exactly this setup):

- `v0.5.0` also produces a non-empty plan or an error → the **image** is the cause, and re-recording is legitimate;
- `v0.5.0` is clean and only the provider under test is not → **compatibility is broken**, and re-recording is
  forbidden. Fix the provider or report the finding.

**Never "fix" a red test by editing `<os>/main.tf`** — not by removing a resource, not by adding attributes which
upstream `v0.5.0` does not know. `record.sh` pins the provider to tag `v0.5.0`, so changing the reference
configuration is the only way to launder a real incompatibility, and that is precisely what is prohibited here. A
configuration change is allowed only together with a documented change of the images, and then for every target.

A move of `system_file` to the plugin framework cannot keep `etag=…` in `source` and therefore makes exactly one
in-place `update` of `system_file.source` expected. Such an exception belongs in the assertion of `compat_test.go`
as an explicit allowlist — never in a re-recorded fixture.

## Notes

- The test does not run `terraform init`. The provider under test is wired in with `dev_overrides` and is not
  published in any registry, so `init` fails to query `neuspaces/system`; Terraform itself recommends skipping
  `init` under development overrides. The reference configuration has no backend and no modules, so `apply`,
  `state replace-provider` and `plan` work without it.
- The test plans with `-out=plan.out` and, on exit code 2, lists the changed resources from
  `show -json plan.out` of that same plan, so the diagnostics always describe the plan which failed the test and not
  a second refresh. Saving a plan file works under development overrides (verified on Terraform 1.6 and 1.12 and on
  OpenTofu 1.10).
- `system_systemd_unit` in the reference configuration sets `status`. With neither `status` nor `enabled` set,
  `v0.5.0` builds an empty shell function for the apply command and fails with `systemd unit resource: unexpected
  error` (`internal/client/systemd_unit.go`), so such a configuration cannot be recorded.
