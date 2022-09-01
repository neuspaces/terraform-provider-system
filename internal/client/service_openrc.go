package client

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/client/openrc"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"regexp"
	"strconv"
	"strings"
)

const ServiceSupervisorOpenRC ServiceSupervisor = "openrc"

func NewOpenRcServiceClient(s system.System) ServiceClient {
	return &openrcServiceClient{
		s: s,
	}
}

type openrcServiceClient struct {
	s system.System
}

var _ ServiceClient = &openrcServiceClient{}

var (
	openrcStatusRegexp = regexp.MustCompile(`^status:(\d+)$`)

	openrcEnabledRegexp = regexp.MustCompile(`^enabled:(\d+)$`)
)

const (
	codeOpenrcServiceNotFound = 16

	codeOpenrcRunlevelNotFound = 17
)

func (c *openrcServiceClient) Get(ctx context.Context, args ServiceGetArgs) (*Service, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { rc-service -q -e '%[1]s' || return %[3]d; [ -d '/etc/runlevels/%[2]s' ] || return %[4]d; { rc-service -q -C '%[1]s' status; echo "status:$?"; }; { [ ! -L '/etc/runlevels/%[2]s/%[1]s' ]; echo "enabled:$?"; }; }; _do;`, args.Name, args.Runlevel, codeOpenrcServiceNotFound, codeOpenrcRunlevelNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, ErrService.Raise(err)
	}

	switch res.ExitCode {
	case codeOpenrcServiceNotFound:
		return nil, ErrServiceNotFound
	case codeOpenrcRunlevelNotFound:
		return nil, ErrServiceRunlevelNotFound
	}

	if res.ExitCode != 0 {
		return nil, ErrServiceUnexpected
	}

	stdoutLines := strings.Split(strings.TrimSpace(string(res.Stdout)), "\n")

	if len(stdoutLines) != 2 || len(res.Stderr) > 0 {
		return nil, ErrServiceUnexpected
	}

	// Parse status
	statusMatch := openrcStatusRegexp.FindStringSubmatch(stdoutLines[0])
	if statusMatch == nil || len(statusMatch) != 2 {
		return nil, ErrServiceUnexpected
	}

	statusCode, err := strconv.Atoi(statusMatch[1])
	if err != nil {
		return nil, ErrServiceUnexpected
	}

	openrcStatus, err := openrc.StatusFromExitCode(statusCode)
	if err != nil {
		return nil, ErrServiceUnexpected
	}

	status := openRcStatusToServiceStatus(openrcStatus)

	// Parse enabled
	enabledMatch := openrcEnabledRegexp.FindStringSubmatch(stdoutLines[1])
	if enabledMatch == nil || len(enabledMatch) != 2 {
		return nil, ErrServiceUnexpected
	}

	enabledCode, err := strconv.Atoi(enabledMatch[1])
	if err != nil || enabledCode < 0 || enabledCode > 1 {
		return nil, ErrServiceUnexpected
	}

	enabled := enabledCode == 1

	svc := &Service{
		Supervisor: ServiceSupervisorOpenRC,
		Name:       args.Name,
		Runlevel:   args.Runlevel,
		Enabled:    to.BoolPtr(enabled),
		Status:     ServiceStatusPtr(status),
	}

	return svc, nil
}

func (c *openrcServiceClient) Apply(ctx context.Context, s Service, opts ...ServiceApplyOption) error {
	// Commands
	var applyCmds []string

	// Activation: Enable/disable in runlevel
	if s.Enabled != nil {
		if *s.Enabled {
			// Command enables the service in the runlevel by creating a symbolic link
			applyCmds = append(applyCmds, fmt.Sprintf(` { [ -L '/etc/runlevels/%[2]s/%[1]s' ] || ln -s '/etc/init.d/%[1]s' '/etc/runlevels/%[2]s/%[1]s'; };`, s.Name, s.Runlevel))
		} else {
			// Command disables the service in the runlevel by removing the symbolic link
			applyCmds = append(applyCmds, fmt.Sprintf(` { [ ! -e '/etc/runlevels/%[2]s/%[1]s' ] || rm -f '/etc/runlevels/%[2]s/%[1]s'; };`, s.Name, s.Runlevel))
		}
	}

	// Status: Start/stop
	if s.Status != nil {
		var rcServiceAction string
		var rcServiceStatusCode int

		switch *s.Status {
		case ServiceStatusStarted:
			rcServiceAction = "start"
			rcServiceStatusCode = 0
		case ServiceStatusStopped:
			rcServiceAction = "stop"
			rcServiceStatusCode = 3
		}

		if rcServiceAction != "" {
			// Command determines the current service status and applies start/stop action if the service is not in the expected status
			applyCmds = append(applyCmds, fmt.Sprintf(` { rc-service -q -C '%[1]s' status; status_rc=$?; [ $status_rc -eq %[2]d ] || rc-service -q '%[1]s' %[3]s || return 1; };`, s.Name, rcServiceStatusCode, rcServiceAction))
		}
	}

	if len(applyCmds) == 0 {
		// Nothing to apply
		return nil
	}

	// Apply changes
	// - all changes are synchronous operations, i.e. Apply will return when all operations are completed
	cmd := NewCommand(fmt.Sprintf(`_do() { rc-service -q -e '%[1]s' || return %[2]d;%[3]s }; _do;`, s.Name, codeOpenrcServiceNotFound, strings.Join(applyCmds, "")))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return err
	}

	switch res.ExitCode {
	case codeOpenrcServiceNotFound:
		return ErrServiceNotFound
	}

	if res.ExitCode != 0 {
		return ErrServiceUnexpected
	}

	return nil
}

func openRcStatusToServiceStatus(s openrc.Status) ServiceStatus {
	switch s {
	case openrc.StatusStopping:
		return ServiceStatusStopping
	case openrc.StatusStarting:
		return ServiceStatusStarting
	case openrc.StatusInactive:
		return ServiceStatusStopped
	case openrc.StatusCrashed:
		return ServiceStatusStopped
	case openrc.StatusStarted:
		return ServiceStatusStarted
	case openrc.StatusStopped:
		return ServiceStatusStopped
	default:
		return ServiceStatusUndefined
	}
}
