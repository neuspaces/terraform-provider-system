package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/neuspaces/terraform-provider-system/internal/client/systemd"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"strings"
)

type SystemdUnit struct {
	Type    string
	Name    string
	Enabled *bool
	Status  *SystemdUnitStatus
}

type SystemdUnitStatus string

func (s SystemdUnitStatus) IsPending() bool {
	return s == SystemdUnitStatusStarting || s == SystemdUnitStatusStopping
}

func SystemdUnitStatusPtr(s SystemdUnitStatus) *SystemdUnitStatus {
	return &s
}

const (
	SystemdUnitStatusStarted SystemdUnitStatus = "started"

	SystemdUnitStatusStarting SystemdUnitStatus = "starting"

	SystemdUnitStatusStopped SystemdUnitStatus = "stopped"

	SystemdUnitStatusStopping SystemdUnitStatus = "stopping"

	SystemdUnitStatusUndefined SystemdUnitStatus = "undefined"
)

type SystemdUnitGetArgs struct {
	Type string
	Name string
}

type SystemdUnitApplyOptions struct {
	restart bool
	reload  bool
}

type SystemdUnitApplyOption func(*SystemdUnitApplyOptions)

// SystemdUnitRestart is an option for Apply to ensure a systemd unit is restarted
func SystemdUnitRestart() SystemdUnitApplyOption {
	return func(o *SystemdUnitApplyOptions) {
		o.restart = true
	}
}

// SystemdUnitReload is an option for Apply to ensure a systemd unit is reloaded
func SystemdUnitReload() SystemdUnitApplyOption {
	return func(o *SystemdUnitApplyOptions) {
		o.reload = true
	}
}

type SystemdUnitClient interface {
	Get(ctx context.Context, args SystemdUnitGetArgs) (*SystemdUnit, error)
	Apply(ctx context.Context, s SystemdUnit, opts ...SystemdUnitApplyOption) error
}

var (
	ErrSystemdUnit = errors.New("systemd unit resource")

	ErrSystemdUnitNotFound = errors.Join(ErrSystemdUnit, errors.New("systemd unit not found"))

	ErrSystemdUnitOperation = errors.Join(ErrSystemdUnit, errors.New("failed systemd unit operation"))

	ErrSystemdUnitUnexpected = errors.Join(ErrSystemdUnit, errors.New("unexpected error"))
)

func NewSystemdUnitClient(s system.System) SystemdUnitClient {
	return &systemdUnitClient{
		s: s,
	}
}

type systemdUnitClient struct {
	s system.System
}

var _ SystemdUnitClient = &systemdUnitClient{}

func (c *systemdUnitClient) Get(ctx context.Context, args SystemdUnitGetArgs) (*SystemdUnit, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { systemctl daemon-reload; systemctl show '%[1]s.%[2]s' --property=LoadState,ActiveState,SubState --plain --no-page; echo "IsEnabled=$(systemctl is-enabled '%[1]s.%[2]s' 2> /dev/null || true)"; }; _do;`, args.Name, args.Type))

	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, errors.Join(ErrSystemdUnit, err)
	}

	if res.ExitCode != 0 {
		return nil, ErrSystemdUnitUnexpected
	}

	// Parse properties
	props, err := godotenv.Parse(bytes.NewReader(res.Stdout))
	if err != nil {
		return nil, errors.Join(ErrSystemdUnitUnexpected, err)
	}

	// Test unit existence
	propLoadState, hasLoadState := props[systemd.PropertyLoadState]
	if !hasLoadState {
		return nil, ErrSystemdUnitUnexpected
	}

	loadState := systemd.LoadState(strings.TrimSpace(propLoadState))

	if loadState == systemd.LoadStateNotFound {
		return nil, ErrSystemdUnitNotFound
	}

	// Status
	propActiveState, hasActiveState := props[systemd.PropertyActiveState]
	if !hasActiveState {
		return nil, ErrSystemdUnitUnexpected
	}

	activeState := systemd.ActiveState(strings.TrimSpace(propActiveState))
	status := systemdActiveStateToSystemdUnitStatus(activeState)

	// Activation: Enabled or disabled?
	// IsEnabled is not a property returned by `systemctl show` but appended by the get command
	propIsEnabled, hasIsEnabled := props["IsEnabled"]
	if !hasIsEnabled {
		return nil, ErrSystemdUnitUnexpected
	}

	isEnabledOutput := systemd.IsEnabledOutput(strings.TrimSpace(propIsEnabled))

	enabled := isEnabledOutput == systemd.Enabled || isEnabledOutput == systemd.EnabledRuntime

	// Unit
	unit := &SystemdUnit{
		Type:    args.Type,
		Name:    args.Name,
		Enabled: to.BoolPtr(enabled),
		Status:  SystemdUnitStatusPtr(status),
	}

	return unit, nil
}

func (c *systemdUnitClient) Apply(ctx context.Context, s SystemdUnit, opts ...SystemdUnitApplyOption) error {
	// Commands
	var applyCmds []string

	// Options
	o := &SystemdUnitApplyOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Activation: Enable/disable
	if s.Enabled != nil {
		if *s.Enabled {
			// Command enables the unit in all targets which are referenced in the [Install] section
			// `systemctl enable $unit` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl enable '%[1]s.%[2]s' --quiet; echo "systemctl_enable_rc=$?"; };`, s.Name, s.Type))
		} else {
			// Command disables the unit
			// `systemctl disable $unit` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl disable '%[1]s.%[2]s' --quiet; echo "systemctl_disable_rc=$?"; };`, s.Name, s.Type))
		}
	}

	// Status: Start/stop
	if s.Status != nil {
		if *s.Status == SystemdUnitStatusStarted {
			// Command starts the unit
			// `systemctl start $unit` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl start '%[1]s.%[2]s' --quiet; echo "systemctl_start_rc=$?"; };`, s.Name, s.Type))

			if o.restart {
				// Restart unit
				// `systemctl restart $unit`
				applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl restart '%[1]s.%[2]s' --quiet; echo "systemctl_restart_rc=$?"; };`, s.Name, s.Type))
			} else if o.reload {
				// Reload unit
				// `systemctl reload $unit`
				applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl reload '%[1]s.%[2]s' --quiet; echo "systemctl_reload_rc=$?"; };`, s.Name, s.Type))
			}
		} else if *s.Status == SystemdUnitStatusStopped {
			// Command stops the unit
			// `systemctl stop $unit` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl stop '%[1]s.%[2]s' --quiet; echo "systemctl_stop_rc=$?"; };`, s.Name, s.Type))
		}
	}

	// Apply changes
	cmd := NewCommand(fmt.Sprintf(`_do() { %[1]s }; _do;`, strings.Join(applyCmds, " ")))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return ErrSystemdUnitUnexpected
	}

	// Parse output properties
	stdoutProps, err := godotenv.Parse(bytes.NewReader(res.Stdout))
	if err != nil {
		return errors.Join(ErrSystemdUnitUnexpected, err)
	}

	if s.Enabled != nil {
		if *s.Enabled {
			if rc := stdoutProps["systemctl_enable_rc"]; rc != "0" {
				return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl enable '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
			}
		} else {
			if rc := stdoutProps["systemctl_disable_rc"]; rc != "0" {
				return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl disable '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
			}
		}
	}

	if s.Status != nil {
		if *s.Status == SystemdUnitStatusStarted {
			if rc := stdoutProps["systemctl_start_rc"]; rc != "0" {
				return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl start '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
			}
		} else if *s.Status == SystemdUnitStatusStopped {
			if rc := stdoutProps["systemctl_stop_rc"]; rc != "0" {
				return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl stop '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
			}
		}
	}

	if rc, ok := stdoutProps["systemctl_restart_rc"]; ok && rc != "0" {
		return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl restart '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
	}

	if rc, ok := stdoutProps["systemctl_reload_rc"]; ok && rc != "0" {
		return errors.Join(ErrSystemdUnitOperation, fmt.Errorf("systemctl reload '%[1]s.%[2]s' returned unexpected exit code %[3]s", s.Name, s.Type, rc))
	}

	return nil
}

func systemdActiveStateToSystemdUnitStatus(s systemd.ActiveState) SystemdUnitStatus {
	switch s {
	case systemd.ActiveStateActive:
		return SystemdUnitStatusStarted
	case systemd.ActiveStateReloading:
		return SystemdUnitStatusUndefined
	case systemd.ActiveStateInactive:
		return SystemdUnitStatusStopped
	case systemd.ActiveStateFailed:
		return SystemdUnitStatusStopped
	case systemd.ActiveStateActivating:
		return SystemdUnitStatusStarted
	case systemd.ActiveStateDeactivating:
		return SystemdUnitStatusStopped
	default:
		return SystemdUnitStatusUndefined
	}
}
