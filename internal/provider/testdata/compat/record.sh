#!/usr/bin/env sh
# Records state.v0.5.0.json for one target with a provider binary built from upstream tag v0.5.0.
#
# Usage: record.sh <alpine|debian|fedora> <ssh_port> <ssh_host_key_file>
#
# Example (acceptance test containers running, repository root as working directory):
#   internal/provider/testdata/compat/record.sh alpine 10022 test/alpine/hostkeys/ssh_host_ecdsa_key.pub
#   internal/provider/testdata/compat/record.sh debian 11022 test/debian/hostkeys/ssh_host_ecdsa_key.pub
#   internal/provider/testdata/compat/record.sh fedora 12022 test/fedora/hostkeys/ssh_host_ecdsa_key.pub
#
# Requires: the docker container of the target running, terraform or tofu (TERRAFORM env, defaults to
# `terraform` in PATH), a go toolchain and network access for `go mod download`.
#
# Connection parameters match internal/provider/acctest.yaml (test containers, test credentials).
# Override with SSH_HOST, SSH_USER and SSH_PASSWORD if the containers run elsewhere.
set -eu

os=$1
port=$2
hostkey_file=$3

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../../.." && pwd)
hostkey_file=$(cd "$(dirname "$hostkey_file")" && pwd)/$(basename "$hostkey_file")

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

TERRAFORM=${TERRAFORM:-terraform}
SSH_HOST=${SSH_HOST:-127.0.0.1}
SSH_USER=${SSH_USER:-root}
SSH_PASSWORD=${SSH_PASSWORD:-'root!'}

# 1. build upstream v0.5.0 from the tag (same code as the released binary)
git -C "$repo" archive v0.5.0 | tar -x -C "$work"
mkdir -p "$work/bin"
(cd "$work" && GOFLAGS=-mod=mod go build -o "$work/bin/terraform-provider-system" .)

# 2. dev_overrides for neuspaces/system -> the v0.5.0 binary
cat > "$work/terraformrc" <<RC
provider_installation {
  dev_overrides { "neuspaces/system" = "$work/bin" }
  direct {}
}
RC

mkdir -p "$work/tf"
cp "$here/$os/main.tf" "$here/$os/hello.txt" "$work/tf/"

cat > "$work/tf/terraform.tfvars.json" <<VARS
{
  "ssh_host": "$SSH_HOST",
  "ssh_port": $port,
  "ssh_user": "$SSH_USER",
  "ssh_password": "$SSH_PASSWORD",
  "ssh_host_key": "$(cut -d' ' -f1,2 < "$hostkey_file")"
}
VARS

# 3. apply with v0.5.0, keep the state, destroy
export TF_CLI_CONFIG_FILE="$work/terraformrc" TF_IN_AUTOMATION=1 TF_INPUT=0
$TERRAFORM -chdir="$work/tf" init -input=false >/dev/null

rc=0
$TERRAFORM -chdir="$work/tf" apply -auto-approve || rc=$?
if [ "$rc" -ne 0 ]; then
    # Do not keep a partial state as fixture, but do remove whatever has been created
    echo "apply failed with exit code $rc; destroying partially created objects" >&2
    $TERRAFORM -chdir="$work/tf" destroy -auto-approve || echo "destroy failed; objects of $os need manual cleanup" >&2
    exit "$rc"
fi

cp "$work/tf/terraform.tfstate" "$here/$os/state.v0.5.0.json"
$TERRAFORM -chdir="$work/tf" destroy -auto-approve

echo "recorded $here/$os/state.v0.5.0.json"
