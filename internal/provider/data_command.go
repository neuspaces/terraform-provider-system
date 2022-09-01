package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/lib/limited"
	"io"
)

const dataCommandName = "system_command"

const (
	dataCommandAttrCommand           = "command"
	dataCommandAttrExitCode          = "exit_code"
	dataCommandAttrStdout            = "stdout"
	dataCommandAttrStderr            = "stderr"
	dataCommandAttrExpect            = "expect"
	dataCommandAttrExpectExitCode    = "exit_code"
	dataCommandAttrExpectStdout      = "stdout"
	dataCommandAttrExpectStdoutLimit = "stdout_limit"
	dataCommandAttrExpectStderr      = "stderr"
	dataCommandAttrExpectStderrLimit = "stderr_limit"
)

const (
	dataCommandStdoutLimitDefault = 65536 // bytes

	dataCommandStderrLimitDefault = 65536 // bytes
)

func dataCommand() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` executes a user-defined command and provides the stdout, stderr, and exit code.", dataCommandName),
		ReadContext: dataCommandRead,
		Schema: map[string]*schema.Schema{
			dataCommandAttrCommand: {
				Description: "Command to execute including arguments. The command is executed in a shell context. Example: `uname -a`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			dataCommandAttrExitCode: {
				Description: "Exit code returned from the command.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataCommandAttrStdout: {
				Description: fmt.Sprintf("Base64 encoded stdout from the command. Captured stdout is limited to a maximum size of %d bytes by default to prevent unindented growth of the terraform state. If necessary, adjust the limit using the attribute `%s`. If the stdout exceeds this limit, the data source fails.", dataCommandStdoutLimitDefault, attrPath{dataCommandAttrExpect, dataCommandAttrExpectStdoutLimit}.String()),
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataCommandAttrStderr: {
				Description: fmt.Sprintf("Base64 encoded stderr from the command. Captured stdout is limited to a maximum size of %d bytes by default to prevent unindented growth of the terraform state. If necessary, adjust the limit using the attribute `%s`. If the stdout exceeds this limit, the data source fails.", dataCommandStdoutLimitDefault, attrPath{dataCommandAttrExpect, dataCommandAttrExpectStderrLimit}.String()),
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataCommandAttrExpect: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						dataCommandAttrExpectExitCode: {
							Description: "Expected exit code from the command. If the exit code returned from the command differs, the data source fails with an error. Defaults to `0`.",
							Type:        schema.TypeInt,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return 0, nil
							},
						},
						dataCommandAttrExpectStdout: {
							Description: "If `true`, the stdout from the command will be captured and provided in output attribute `stdout`. Defaults to `true`.",
							Type:        schema.TypeBool,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return true, nil
							},
						},
						dataCommandAttrExpectStdoutLimit: {
							Description:  fmt.Sprintf("Maximum bytes read from stdout of the command. Define a reasonable limit to prevent unindented growth of the terraform state. Defaults to `%d`.", dataCommandStdoutLimitDefault),
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							DefaultFunc: func() (interface{}, error) {
								return dataCommandStdoutLimitDefault, nil
							},
						},
						dataCommandAttrExpectStderr: {
							Description: "If `true`, the stderr from the command will be captured and provided in output attribute `stderr`. Defaults to `true`.",
							Type:        schema.TypeBool,
							Optional:    true,
							DefaultFunc: func() (interface{}, error) {
								return true, nil
							},
						},
						dataCommandAttrExpectStderrLimit: {
							Description:  fmt.Sprintf("Maximum bytes read from stderr of the command. Define a reasonable limit to prevent unindented growth of the terraform state. Defaults to `%d`.", dataCommandStderrLimitDefault),
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							DefaultFunc: func() (interface{}, error) {
								return dataCommandStderrLimitDefault, nil
							},
						},
					},
				},
			},
		},
	}
}

type dataCommandExpect struct {
	ExitCode    int
	Stdout      bool
	StdoutLimit int
	Stderr      bool
	StderrLimit int
}

func expandDataCommandExpect(v interface{}) (*dataCommandExpect, error) {
	d, err := expandListSingle(v)
	if err != nil {
		return nil, err
	}

	e := &dataCommandExpect{
		ExitCode:    d[dataCommandAttrExitCode].(int),
		Stdout:      d[dataCommandAttrStdout].(bool),
		StdoutLimit: d[dataCommandAttrExpectStdoutLimit].(int),
		Stderr:      d[dataCommandAttrStderr].(bool),
		StderrLimit: d[dataCommandAttrExpectStderrLimit].(int),
	}

	return e, nil
}

func defaultDataCommandExpect() *dataCommandExpect {
	return &dataCommandExpect{
		ExitCode:    0,
		Stdout:      true,
		StdoutLimit: dataCommandStdoutLimitDefault,
		Stderr:      true,
		StderrLimit: dataCommandStderrLimitDefault,
	}
}

func dataCommandRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	commandString := d.Get(dataCommandAttrCommand).(string)

	var expect *dataCommandExpect

	if expectV, expectOk := d.GetOk(dataCommandAttrExpect); expectOk {
		e, err := expandDataCommandExpect(expectV)
		if err != nil {
			return diag.FromErr(err)
		}
		expect = e
	} else {
		expect = defaultDataCommandExpect()
	}

	// Prepare command options
	var commandOptions []client.ExecuteCommandOption

	if expect.Stdout {
		commandOptions = append(commandOptions, client.WithStdoutFunc(func(w io.Writer) io.Writer {
			return limited.NewWriter(w, int64(expect.StdoutLimit))
		}))
	}

	if expect.Stderr {
		commandOptions = append(commandOptions, client.WithStderrFunc(func(w io.Writer) io.Writer {
			return limited.NewWriter(w, int64(expect.StderrLimit))
		}))
	}

	// Execute command
	command := client.NewCommand(commandString)
	result, err := client.ExecuteCommandWithOptions(ctx, p.System, command, commandOptions...)
	if err != nil {
		if err.Error() == "short write" {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "stdout or stderr exceeded limit",
				},
			}
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("unexpected error: %s", err.Error()),
					Detail:   err.Error(),
				},
			}
		}
	}

	// Exit code should match expected exit code
	if result.ExitCode != expect.ExitCode {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("expected exit code %d, got exit code %d", expect.ExitCode, result.ExitCode),
			},
		}
	}

	// Terraform requires an id: Use the hex encoded sha1 sum of a string concat of all attributes
	id, err := dataIdFromAttrValues(commandString, result.ExitCode, result.StdoutString(), result.StderrString())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	_ = d.Set(dataCommandAttrExitCode, result.ExitCode)
	_ = d.Set(dataCommandAttrStdout, base64.StdEncoding.EncodeToString(result.Stdout))
	_ = d.Set(dataCommandAttrStderr, base64.StdEncoding.EncodeToString(result.Stderr))

	return nil
}
