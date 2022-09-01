package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/neuspaces/terraform-provider-system/internal/system"
)

// GetOsReleaseInfo returns an osrelease.Info based on the /etc/os-release of the provided system.System
func GetOsReleaseInfo(ctx context.Context, s system.System) (*osrelease.Info, error) {
	catCmd := &CatCommand{Path: "/etc/os-release"}
	res, err := ExecuteCommand(ctx, s, catCmd)
	if err != nil {
		return nil, err
	}

	if res.ExitCode != 0 {
		return nil, fmt.Errorf("non-zero exit code")
	}

	info, err := osrelease.Parse(bytes.NewReader(res.Stdout))
	if err != nil || info == nil {
		return nil, err
	}

	return info, nil
}
