package systemd

type IsEnabledOutput string

const (
	Enabled IsEnabledOutput = "enabled"

	EnabledRuntime IsEnabledOutput = "enabled-runtime"

	Linked IsEnabledOutput = "linked"

	LinkedRuntime IsEnabledOutput = "linked-runtime"

	Masked IsEnabledOutput = "masked"

	MaskedRuntime IsEnabledOutput = "masked-runtime"

	Static IsEnabledOutput = "static"

	Indirect IsEnabledOutput = "indirect"

	Disabled IsEnabledOutput = "disabled"

	Generated IsEnabledOutput = "generated"

	Transient IsEnabledOutput = "transient"
)

// IsEnabledOutputs provides a list of valid outputs from `systemctl is-enabled UNIT…`
// https://www.freedesktop.org/software/systemd/man/systemctl.html#is-enabled%20UNIT%E2%80%A6
var IsEnabledOutputs = []IsEnabledOutput{
	Enabled,
	EnabledRuntime,
	Linked,
	LinkedRuntime,
	Masked,
	MaskedRuntime,
	Static,
	Indirect,
	Disabled,
	Generated,
	Transient,
}
