package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuspaces/terraform-provider-system/internal/acctest"
)

const (
	// compatUpstreamAddr is the provider address which upstream v0.5.0 wrote into the state fixtures
	compatUpstreamAddr = "registry.terraform.io/neuspaces/system"

	// compatForkNamespace is the namespace and type of the fork, without the registry host
	compatForkNamespace = "dorokhovich-com/system"

	// compatUpstreamSource is the required_providers source of upstream v0.5.0 in the reference configuration
	compatUpstreamSource = `source = "neuspaces/system"`
)

// TestAccCompat_upstream_state_empty_plan verifies the drop-in contract (spec section 7.3): a state file written by
// upstream v0.5.0 for the reference configuration yields an empty plan with this provider after
// `terraform state replace-provider`.
//
// Per target, the test
//  1. creates the remote objects of the reference configuration with the provider under test from an empty state
//     (directory bootstrap),
//  2. drops the recorded upstream state into a second directory (directory migrate) and moves it to the provider
//     address of the fork, which is the second of the two migration operations,
//  3. requires `plan -detailed-exitcode` to exit 0; exit code 2 (changes present) and exit code 1 (error) fail the
//     test, and
//  4. destroys the objects through the migrated state, which exercises Delete on the migrated state.
//
// The provider under test is wired in with dev_overrides, so the provider address compiled into the binary is
// irrelevant and the registry host comes from the CLI (Terraform or OpenTofu).
//
// Fixtures live in testdata/compat/<os>/ and are recorded with testdata/compat/record.sh. When the plan is not empty,
// the fixture is not to be re-recorded: see testdata/compat/README.md.
func TestAccCompat_upstream_state_empty_plan(t *testing.T) {
	if os.Getenv(acctest.EnvTfAcc) == "" {
		t.Skipf("acceptance test skipped because %s is not set", acctest.EnvTfAcc)
	}

	cli := compatCLI(t)
	providerDir := compatBuildProvider(t)
	env := compatEnv(t, providerDir)

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		fixture := filepath.Join("testdata", "compat", target.Os.Id)
		state := filepath.Join(fixture, "state.v0.5.0.json")
		if _, err := os.Stat(state); err != nil {
			t.Fatalf("missing fixture %s: record it with testdata/compat/record.sh", state)
		}

		bootstrap := compatWorkdir(t, fixture, target)
		migrate := compatWorkdir(t, fixture, target)

		// Remove the remote objects at the end of the test, preferably through the migrated state
		t.Cleanup(func() {
			if _, err := os.Stat(filepath.Join(migrate, "terraform.tfstate")); err == nil {
				if out, err := compatExec(cli.path, migrate, env, "destroy", "-auto-approve", "-input=false"); err != nil {
					t.Errorf("destroy from migrated state failed: %v\n%s", err, out)
				} else {
					return
				}
			}

			if out, err := compatExec(cli.path, bootstrap, env, "destroy", "-auto-approve", "-input=false"); err != nil {
				t.Errorf("destroy from bootstrap state failed: %v\n%s", err, out)
			}
		})

		// 1. materialize the remote objects with the provider under test from an empty state
		//
		// `init` is intentionally not run: the provider under test is wired in with dev_overrides and is not
		// published in a registry, so `init` fails to query it ("Skip terraform init when using provider
		// development overrides"). The configuration has no backend and no modules, hence `apply` and `plan`
		// work without it.
		compatRun(t, cli.path, bootstrap, env, "apply", "-auto-approve", "-input=false")

		// 2. drop in the upstream state and move it to the provider address of the fork
		compatCopy(t, state, filepath.Join(migrate, "terraform.tfstate"))
		compatRun(t, cli.path, migrate, env, "state", "replace-provider", "-auto-approve", compatUpstreamAddr, cli.host+"/"+compatForkNamespace)

		// 3. the plan must be empty
		out, err := compatExec(cli.path, migrate, env, "plan", "-detailed-exitcode", "-input=false", "-no-color")

		var exitErr *exec.ExitError
		switch {
		case err == nil:
			// exit code 0: no changes
			return
		case errors.As(err, &exitErr) && exitErr.ExitCode() == 2:
			t.Fatalf("plan is not empty after migrating the upstream v0.5.0 state:\n%s\n%s", compatPlannedChanges(cli, migrate, env), out)
		default:
			t.Fatalf("plan failed: %v\n%s", err, out)
		}
	})
}

type compatCLIInfo struct {
	// path is the path of the Terraform or OpenTofu executable
	path string

	// host is the registry host of the CLI, either registry.terraform.io or registry.opentofu.org
	host string
}

// compatCLI returns the CLI to run the test with, from TF_ACC_TERRAFORM_PATH or `terraform` in the path
func compatCLI(t *testing.T) compatCLIInfo {
	p := os.Getenv(acctest.EnvTfAccTerraformPath)
	if p == "" {
		var err error
		if p, err = exec.LookPath("terraform"); err != nil {
			t.Fatalf("terraform not found in path: set %s", acctest.EnvTfAccTerraformPath)
		}
	}

	v, err := exec.Command(p, "version").Output()
	if err != nil {
		t.Fatalf("%s version: %v", p, err)
	}

	host := "registry.terraform.io"
	if strings.Contains(string(v), "OpenTofu") {
		host = "registry.opentofu.org"
	}

	return compatCLIInfo{path: p, host: host}
}

// compatBuildProvider builds the provider under test into a temporary directory which is used as dev_overrides target
func compatBuildProvider(t *testing.T) string {
	dir := t.TempDir()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found at %s", root)
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "terraform-provider-system"), ".")
	cmd.Dir = root

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	return dir
}

// compatEnv returns the environment of the CLI invocations, including a CLI configuration with dev_overrides for the
// provider under test
func compatEnv(t *testing.T, providerDir string) []string {
	rc := filepath.Join(t.TempDir(), "terraformrc")

	content := fmt.Sprintf("provider_installation {\n  dev_overrides { %q = %q }\n  direct {}\n}\n", compatForkNamespace, providerDir)
	if err := os.WriteFile(rc, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	env := []string{
		"TF_CLI_CONFIG_FILE=" + rc,
		"TF_IN_AUTOMATION=1",
		"TF_INPUT=0",
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
	}

	if v := os.Getenv(acctest.EnvTfLog); v != "" {
		env = append(env, acctest.EnvTfLog+"="+v)
	}

	return env
}

// compatWorkdir copies the reference configuration of the fixture into a temporary directory, rewrites the provider
// source to the fork, which is the first of the two migration operations, and writes the connection variables
func compatWorkdir(t *testing.T, fixture string, target acctest.Target) string {
	dir := t.TempDir()

	main, err := os.ReadFile(filepath.Join(fixture, "main.tf"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(main), compatUpstreamSource) {
		t.Fatalf("expected %s in %s", compatUpstreamSource, filepath.Join(fixture, "main.tf"))
	}

	main = []byte(strings.Replace(string(main), compatUpstreamSource, fmt.Sprintf("source = %q", compatForkNamespace), 1))

	if err := os.WriteFile(filepath.Join(dir, "main.tf"), main, 0o644); err != nil {
		t.Fatal(err)
	}

	compatCopy(t, filepath.Join(fixture, "hello.txt"), filepath.Join(dir, "hello.txt"))

	ssh := target.Configs.Default().Ssh

	vars, err := json.Marshal(map[string]interface{}{
		"ssh_host":     ssh.Host,
		"ssh_port":     ssh.Port,
		"ssh_user":     ssh.User,
		"ssh_password": ssh.Password,
		"ssh_host_key": strings.TrimSpace(ssh.HostKey),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "terraform.tfvars.json"), vars, 0o600); err != nil {
		t.Fatal(err)
	}

	return dir
}

func compatCopy(t *testing.T, src string, dst string) {
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func compatExec(cli string, dir string, env []string, args ...string) (string, error) {
	cmd := exec.Command(cli, args...)
	cmd.Dir = dir
	cmd.Env = env

	out, err := cmd.CombinedOutput()

	return string(out), err
}

func compatRun(t *testing.T, cli string, dir string, env []string, args ...string) string {
	out, err := compatExec(cli, dir, env, args...)
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", filepath.Base(cli), strings.Join(args, " "), err, out)
	}

	return out
}

// compatPlannedChanges lists the planned changes of a plan which are not a no-op.
//
// A plan file cannot be used because the CLI refuses to save a plan while provider development overrides are in
// effect, hence the machine readable output of a second plan.
//
// In v0.7.0, when `source` is updated in place instead of forcing a new resource, exactly one `update` of
// system_file.source is expected and is to be allowed here (owner decision 1 of the phase 1 plan).
func compatPlannedChanges(cli compatCLIInfo, dir string, env []string) string {
	out, _ := compatExec(cli.path, dir, env, "plan", "-input=false", "-no-color", "-json")

	var b strings.Builder

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var msg struct {
			Type   string `json:"type"`
			Change struct {
				Resource struct {
					Addr string `json:"addr"`
				} `json:"resource"`
				Action string `json:"action"`
			} `json:"change"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if msg.Type != "planned_change" || msg.Change.Action == "noop" {
			continue
		}

		fmt.Fprintf(&b, "%s: %s\n", msg.Change.Resource.Addr, msg.Change.Action)
	}

	if b.Len() == 0 {
		return "no planned changes reported by `plan -json`"
	}

	return b.String()
}
