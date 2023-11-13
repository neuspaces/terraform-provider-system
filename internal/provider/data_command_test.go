package provider_test

import (
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/neuspaces/terraform-provider-system/internal/acctest"
	"github.com/neuspaces/terraform-provider-system/internal/acctest/tfbuild"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/heredoc"
	"github.com/neuspaces/terraform-provider-system/internal/provider"
	"regexp"
	"strings"
	"testing"
)

func TestAccDataCommand(t *testing.T) {
	type testCase struct {
		Desc string

		ResourceAddr string
		ResourceHcl  string

		ExpectExitCode string
		ExpectStdout   string
		ExpectStderr   string

		ExpectErr      bool
		ExpectErrRegex string
	}

	tcs := []testCase{
		{
			Desc: "uname",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "uname"
				}
			`),

			ExpectExitCode: "0",
			ExpectStdout:   base64.StdEncoding.EncodeToString([]byte("Linux\n")),
			ExpectStderr:   "",
		},
		{
			Desc: "whoami; uname",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "uname; whoami"
				}
			`),

			ExpectExitCode: "0",
			ExpectStdout:   base64.StdEncoding.EncodeToString([]byte("Linux\nroot\n")),
			ExpectStderr:   "",
		},
		{
			Desc: "uname 1>&2",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "uname 1>&2"
				}
			`),

			ExpectExitCode: "0",
			ExpectStdout:   "",
			ExpectStderr:   base64.StdEncoding.EncodeToString([]byte("Linux\n")),
		},
		{
			Desc: "expected empty stdout",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "uname"
					
					expect {
						stdout = false
					}
				}
			`),

			ExpectExitCode: "0",
			ExpectStdout:   "",
			ExpectStderr:   "",
		},
		{
			Desc: "expected empty stderr",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "uname 1>&2"
					
					expect {
						stderr = false
					}
				}
			`),

			ExpectExitCode: "0",
			ExpectStdout:   "",
			ExpectStderr:   "",
		},
		{
			Desc: "expected non-zero exit code",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "exit 42"
					
					expect {
						exit_code = 42
					}
				}
			`),

			ExpectExitCode: "42",
			ExpectStdout:   "",
			ExpectStderr:   "",
		},
		{
			Desc: "unexpected non-zero exit code",

			ResourceAddr: "data.system_command.text",
			ResourceHcl: heredoc.String(`
				data "system_command" "test" {
					command = "exit 1"
				}
			`),

			ExpectErr:      true,
			ExpectErrRegex: `expected exit code 0, got exit code 1`,
		},
		//{
		//	Desc: "exceeding stdout limit",
		//
		//	ResourceAddr: "data.system_command.text",
		//	ResourceHcl: heredoc.String(`
		//		data "system_command" "test" {
		//			command = "head -c 65537 /dev/random"
		//		}
		//	`),
		//
		//	ExpectErr:      true,
		//	ExpectErrRegex: `stdout or stderr exceeded limit`,
		//},
		//{
		//	Desc: "exceeding stderr limit",
		//
		//	ResourceAddr: "data.system_command.text",
		//	ResourceHcl: heredoc.String(`
		//		data "system_command" "test" {
		//			command = "head -c 65537 /dev/random 1>&2"
		//		}
		//	`),
		//
		//	ExpectErr:      true,
		//	ExpectErrRegex: `stdout or stderr exceeded limit`,
		//},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			acctest.Current().Targets.Foreach(t, func(t *testing.T, target acctest.Target) {
				t.Parallel()

				targetConfig := target.Configs.Default()

				resourceConfig := strings.TrimSpace(tfbuild.FileString(tfbuild.File(
					acctest.ProviderConfigBlock(targetConfig),
				))) + "\n\n" + strings.TrimSpace(tc.ResourceHcl)

				if !tc.ExpectErr {
					resource.Test(t, resource.TestCase{
						ProviderFactories: acctest.ProviderFactories(),
						Steps: []resource.TestStep{
							{
								Config: resourceConfig,
								Check: resource.ComposeTestCheckFunc(
									provider.TestLogResourceAttr(t, "data.system_command.test"),
									resource.TestCheckResourceAttrSet("data.system_command.test", "id"),
									resource.TestCheckResourceAttr("data.system_command.test", "exit_code", tc.ExpectExitCode),
									resource.TestCheckResourceAttr("data.system_command.test", "stdout", tc.ExpectStdout),
									resource.TestCheckResourceAttr("data.system_command.test", "stderr", tc.ExpectStderr),
								),
							},
						},
					})
				} else {
					resource.Test(t, resource.TestCase{
						ProviderFactories: acctest.ProviderFactories(),
						Steps: []resource.TestStep{
							{
								Config:      resourceConfig,
								ExpectError: regexp.MustCompile(tc.ExpectErrRegex),
							},
						},
					})
				}
			})
		})
	}
}
