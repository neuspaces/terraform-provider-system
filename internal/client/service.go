package client

import (
	"context"
	"errors"
)

type Service struct {
	Supervisor ServiceSupervisor
	Name       string
	Target     string
	Runlevel   string
	Enabled    *bool
	Status     *ServiceStatus
}

type ServiceSupervisor string

type ServiceStatus string

func (s ServiceStatus) IsPending() bool {
	return s == ServiceStatusStarting || s == ServiceStatusStopping
}

func ServiceStatusPtr(s ServiceStatus) *ServiceStatus {
	return &s
}

const (
	ServiceStatusStarted ServiceStatus = "started"

	ServiceStatusStarting ServiceStatus = "starting"

	ServiceStatusStopped ServiceStatus = "stopped"

	ServiceStatusStopping ServiceStatus = "stopping"

	//ServiceStatusPending ServiceStatus = "pending"

	ServiceStatusUndefined ServiceStatus = "undefined"
)

type ServiceGetArgs struct {
	Name     string
	Runlevel string
}

type ServiceApplyOptions struct {
	restart bool
	reload  bool
}

type ServiceApplyOption func(*ServiceApplyOptions)

// ServiceRestart is an option for Apply to ensure a service is restarted
func ServiceRestart() ServiceApplyOption {
	return func(o *ServiceApplyOptions) {
		o.restart = true
	}
}

// ServiceReload is an option for Apply to ensure a service is reloaded
func ServiceReload() ServiceApplyOption {
	return func(o *ServiceApplyOptions) {
		o.reload = true
	}
}

type ServiceClient interface {
	Get(ctx context.Context, args ServiceGetArgs) (*Service, error)
	Apply(ctx context.Context, s Service, opts ...ServiceApplyOption) error
}

var (
	ErrService = errors.New("service resource")

	ErrServiceNotFound = errors.Join(ErrService, errors.New("service not found"))

	ErrServiceRunlevelNotFound = errors.Join(ErrService, errors.New("runlevel not found"))

	ErrServiceOperation = errors.Join(ErrService, errors.New("failed service operation"))

	ErrServiceUnexpected = errors.Join(ErrService, errors.New("unexpected error"))
)
