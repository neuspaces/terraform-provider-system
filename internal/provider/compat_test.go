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
	// compatV050Addr is the provider address which v0.5.0 wrote into the state fixtures
	compatV050Addr = "registry.terraform.io/neuspaces/system"

	// compatRegistryHost is the registry host recorded in the state fixtures
	compatRegistryHost = "registry.terraform.io"

	// compatNamespace is the namespace and type of the provider, without the registry host
	compatNamespace = "neuspaces/system"

	// compatSource is the required_providers source in the reference configuration
	compatSource = `source = "neuspaces/system"`
)

// TestAccCompat_v050_state_empty_plan verifies that a state file written by the released v0.5.0 provider for the
// reference configuration yields an empty plan with the provider under test. It guards every change in this
// repository against surprising existing users with a plan they did not ask for.
//
// Per target, the test
//  1. creates the remote objects of the reference configuration with the provider under test from an empty state
//     (directory bootstrap),
//  2. drops the recorded v0.5.0 state into a second directory (directory migrate),
//  3. requires `plan -detailed-exitcode` to exit 0; exit code 2 (changes present) and exit code 1 (error) fail the
//     test, and
//  4. destroys the objects through that state, which exercises Delete on a state written by v0.5.0.
//
// The provider under test is wired in with dev_overrides, so the provider address compiled into the binary is
// irrelevant. The fixtures are recorded with Terraform, so their state refers to registry.terraform.io; under
// OpenTofu the state is moved to the OpenTofu registry host first, which is a CLI difference and not a change of
// the provider.
//
// Fixtures live in testdata/compat/<os>/ and are recorded with testdata/compat/record.sh. When the plan is not empty,
// the fixture is not to be re-recorded: see testdata/compat/README.md.
func TestAccCompat_v050_state_empty_plan(t *testing.T) {
	if os.Getenv(acctest.EnvTfAcc) == "" {
		t.Skipf("acceptance test skipped because %s is not set", acctest.EnvTfAcc)
	}

	cli := compatCLI(t)
	providerDir := compatBuildProvider(t)
	env := compatEnv(t, providerDir)

	// Without this guard the test reports success without asserting anything, for example when a target of
	// TF_ACC_PROVIDER_SYSTEM_TARGETS is misspelled and no target is selected at all
	if len(acctest.Current().Targets) == 0 {
		t.Fatal("no acceptance targets selected")
	}

	acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
		fixture := filepath.Join("testdata", "compat", target.Os.Id)
		state := filepath.Join(fixture, "state.v0.5.0.json")
		if _, err := os.Stat(state); err != nil {
			t.Fatalf("missing fixture %s: record it with testdata/compat/record.sh", state)
		}

		bootstrap := compatWorkdir(t, fixture, target)
		migrate := compatWorkdir(t, fixture, target)

		// usable is true once the state in the migrate directory refers to a provider address which the CLI
		// resolves to the provider under test; before that a destroy from that directory cannot work
		usable := false

		// Remove the remote objects at the end of the test, preferably through the migrated state
		t.Cleanup(func() {
			if usable {
				if out, err := compatExec(cli.path, migrate, env, "destroy", "-auto-approve", "-input=false"); err != nil {
					t.Errorf("destroy from the v0.5.0 state failed: %v\n%s", err, out)
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

		// 2. drop in the state recorded with v0.5.0
		//
		// The fixtures are recorded with Terraform. Under OpenTofu the provider address in the state carries the
		// wrong registry host, which `state replace-provider` corrects; the namespace and type stay the same.
		compatCopy(t, state, filepath.Join(migrate, "terraform.tfstate"))

		if cli.host != compatRegistryHost {
			compatRun(t, cli.path, migrate, env, "state", "replace-provider", "-auto-approve", compatV050Addr, cli.host+"/"+compatNamespace)
		}

		usable = true

		// 3. the plan must be empty
		out, err := compatExec(cli.path, migrate, env, "plan", "-detailed-exitcode", "-input=false", "-no-color", "-out=plan.out")

		var exitErr *exec.ExitError
		switch {
		case err == nil:
			// exit code 0: no changes
			return
		case errors.As(err, &exitErr) && exitErr.ExitCode() == 2:
			// exit code 2: changes present; report them from the very plan which returned exit code 2
			show, showErr := compatExec(cli.path, migrate, env, "show", "-json", "plan.out")
			if showErr != nil {
				t.Fatalf("plan is not empty for the state written by v0.5.0, and `show -json plan.out` failed: %v\n%s\n%s", showErr, show, out)
			}

			t.Fatalf("plan is not empty for the state written by v0.5.0:\n%s\n%s", compatSummarizeChanges(show), out)
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

	content := fmt.Sprintf("provider_installation {\n  dev_overrides { %q = %q }\n  direct {}\n}\n", compatNamespace, providerDir)
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

// compatWorkdir copies the reference configuration of the fixture into a temporary directory unchanged and writes the
// connection variables
func compatWorkdir(t *testing.T, fixture string, target acctest.Target) string {
	dir := t.TempDir()

	main, err := os.ReadFile(filepath.Join(fixture, "main.tf"))
	if err != nil {
		t.Fatal(err)
	}

	// Guards that the fixture is the reference configuration and that dev_overrides applies to the source it declares
	if !strings.Contains(string(main), compatSource) {
		t.Fatalf("expected %s in %s", compatSource, filepath.Join(fixture, "main.tf"))
	}

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

// compatSummarizeChanges lists the resource changes of a `show -json` plan which are not a no-op
func compatSummarizeChanges(showJSON string) string {
	var plan struct {
		ResourceChanges []struct {
			Address string `json:"address"`
			Change  struct {
				Actions []string `json:"actions"`
			} `json:"change"`
		} `json:"resource_changes"`
	}

	if err := json.Unmarshal([]byte(showJSON), &plan); err != nil {
		return showJSON
	}

	var b strings.Builder

	for _, rc := range plan.ResourceChanges {
		if len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "no-op" {
			continue
		}

		fmt.Fprintf(&b, "%s: %s\n", rc.Address, strings.Join(rc.Change.Actions, ","))
	}

	if b.Len() == 0 {
		return "no changed resources in `show -json plan.out`"
	}

	return b.String()
}
