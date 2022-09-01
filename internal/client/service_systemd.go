package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/neuspaces/terraform-provider-system/internal/client/systemd"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"strings"
)

const ServiceSupervisorSystemd ServiceSupervisor = "systemd"

func NewSystemdServiceClient(s system.System) ServiceClient {
	return &systemdServiceClient{
		s: s,
	}
}

type systemdServiceClient struct {
	s system.System
}

var _ ServiceClient = &systemdServiceClient{}

func (c *systemdServiceClient) Get(ctx context.Context, args ServiceGetArgs) (*Service, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { systemctl daemon-reload; systemctl show '%[1]s.service' --property=LoadState,ActiveState,SubState --plain --no-page; echo "IsEnabled=$(systemctl is-enabled '%[1]s.service' 2> /dev/null || true)"; }; _do;`, args.Name))

	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, ErrService.Raise(err)
	}

	if res.ExitCode != 0 {
		return nil, ErrServiceUnexpected
	}

	// Parse properties
	props, err := godotenv.Parse(bytes.NewReader(res.Stdout))
	if err != nil {
		return nil, ErrServiceUnexpected.Raise(err)
	}

	// Test service unit existence
	propLoadState, hasLoadState := props[systemd.PropertyLoadState]
	if !hasLoadState {
		return nil, ErrServiceUnexpected
	}

	loadState := systemd.LoadState(strings.TrimSpace(propLoadState))

	if loadState == systemd.LoadStateNotFound {
		return nil, ErrServiceNotFound
	}

	// Determine masked
	// masked := loadState == systemd.LoadStateMasked

	// Status
	propActiveState, hasActiveState := props[systemd.PropertyActiveState]
	if !hasActiveState {
		return nil, ErrServiceUnexpected
	}

	activeState := systemd.ActiveState(strings.TrimSpace(propActiveState))
	status := systemdActiveStateToServiceStatus(activeState)

	// Activation: Enabled or disabled?
	// IsEnabled is not a property returned by `systemctl show` but appended by the get command
	propIsEnabled, hasIsEnabled := props["IsEnabled"]
	if !hasIsEnabled {
		return nil, ErrServiceUnexpected
	}

	isEnabledOutput := systemd.IsEnabledOutput(strings.TrimSpace(propIsEnabled))

	enabled := isEnabledOutput == systemd.Enabled || isEnabledOutput == systemd.EnabledRuntime

	// Service
	svc := &Service{
		Supervisor: ServiceSupervisorSystemd,
		Name:       args.Name,
		Target:     args.Runlevel,
		Enabled:    to.BoolPtr(enabled),
		Status:     ServiceStatusPtr(status),
	}

	return svc, nil
}

func (c *systemdServiceClient) Apply(ctx context.Context, s Service, opts ...ServiceApplyOption) error {
	// Commands
	var applyCmds []string

	// Activation: Enable/disable
	if s.Enabled != nil {
		if *s.Enabled {
			// Command enables the service unit in all targets which are references in the [Install] section
			// `systemctl enable [unit]` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl enable '%[1]s.service' --quiet; echo "systemctl_enable_rc=$?"; };`, s.Name))
		} else {
			// Command disables the service unit
			// `systemctl disable [unit]` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl disable '%[1]s.service' --quiet; echo "systemctl_disable_rc=$?"; };`, s.Name))
		}
	}

	// Status: Start/stop
	if s.Status != nil {
		if *s.Status == ServiceStatusStarted {
			// Command starts the service unit
			// `systemctl start [unit]` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl start '%[1]s.service' --quiet; echo "systemctl_start_rc=$?"; };`, s.Name))
		} else if *s.Status == ServiceStatusStopped {
			// Command stops the service unit
			// `systemctl stop [unit]` is idempotent
			applyCmds = append(applyCmds, fmt.Sprintf(`{ systemctl stop '%[1]s.service' --quiet; echo "systemctl_stop_rc=$?"; };`, s.Name))
		}
	}

	// Apply changes
	cmd := NewCommand(fmt.Sprintf(`_do() { %[1]s }; _do;`, strings.Join(applyCmds, " ")))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return ErrServiceUnexpected
	}

	// Parse output properties
	stdoutProps, err := godotenv.Parse(bytes.NewReader(res.Stdout))
	if err != nil {
		return ErrServiceUnexpected.Raise(err)
	}

	if s.Enabled != nil {
		if *s.Enabled {
			if rc := stdoutProps["systemctl_enable_rc"]; rc != "0" {
				return ErrServiceOperation.Raise(fmt.Errorf("systemctl enable '%[1]s.service' returned unexpected exit code %[2]s", s.Name, rc))
			}
		} else {
			if rc := stdoutProps["systemctl_disable_rc"]; rc != "0" {
				return ErrServiceOperation.Raise(fmt.Errorf("systemctl disable '%[1]s.service' returned unexpected exit code %[2]s", s.Name, rc))
			}
		}
	}

	if s.Status != nil {
		if *s.Status == ServiceStatusStarted {
			if rc := stdoutProps["systemctl_start_rc"]; rc != "0" {
				return ErrServiceOperation.Raise(fmt.Errorf("systemctl start '%[1]s.service' returned unexpected exit code %[2]s", s.Name, rc))
			}
		} else if *s.Status == ServiceStatusStopped {
			if rc := stdoutProps["systemctl_stop_rc"]; rc != "0" {
				return ErrServiceOperation.Raise(fmt.Errorf("systemctl stop '%[1]s.service' returned unexpected exit code %[2]s", s.Name, rc))
			}
		}
	}

	return nil
}

func systemdActiveStateToServiceStatus(s systemd.ActiveState) ServiceStatus {
	switch s {
	case systemd.ActiveStateActive:
		return ServiceStatusStarted
	case systemd.ActiveStateReloading:
		return ServiceStatusUndefined
	case systemd.ActiveStateInactive:
		return ServiceStatusStopped
	case systemd.ActiveStateFailed:
		return ServiceStatusStopped
	case systemd.ActiveStateActivating:
		return ServiceStatusStarted
	case systemd.ActiveStateDeactivating:
		return ServiceStatusStopped
	default:
		return ServiceStatusUndefined
	}
}
