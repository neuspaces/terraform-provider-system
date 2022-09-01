package openrc

import "fmt"

type Status string

const (
	StatusStopping Status = "stopping"

	StatusStarting Status = "starting"

	StatusInactive Status = "inactive"

	StatusCrashed Status = "crashed"

	StatusStarted Status = "started"

	StatusStopped Status = "stopped"
)

// StatusFromExitCode returns the Status of an OpenRC service from the exit code returned from an OpenRC service script `rc-service [name] status`.
// Reference: https://github.com/OpenRC/openrc/blob/63db2d99e730547339d1bdd28e8437999c380cae/sh/openrc-run.sh.in#L125
func StatusFromExitCode(exitCode int) (Status, error) {
	switch exitCode {
	case 4:
		return StatusStopping, nil
	case 8:
		return StatusStarting, nil
	case 16:
		return StatusInactive, nil
	case 32:
		return StatusCrashed, nil
	case 0:
		return StatusStarted, nil
	case 3:
		return StatusStopped, nil
	default:
		return "", fmt.Errorf("invalid openrc service status code: %d", exitCode)
	}
}
