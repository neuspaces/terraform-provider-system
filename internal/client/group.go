package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"strconv"
	"strings"
)

type Group struct {
	Gid    int
	Name   string
	System bool
}

type GroupClient interface {
	Get(ctx context.Context, gid int) (*Group, error)
	Create(ctx context.Context, group Group) (int, error)
	Update(ctx context.Context, group Group) error
	Delete(ctx context.Context, gid int) error
}

func NewGroupClient(s system.System) GroupClient {
	return &groupClient{
		s: s,
	}
}

var (
	ErrGroup = errors.New("group resource")

	ErrGroupNotFound = errors.Join(ErrGroup, errors.New("group not found"))

	ErrGroupNameExists = errors.Join(ErrGroup, errors.New("group name exists"))

	ErrGroupGidExists = errors.Join(ErrGroup, errors.New("group gid exists"))

	ErrGroupUnexpected = errors.Join(ErrGroup, errors.New("unexpected error"))
)

const (
	codeGroupUnexpected = 1

	codeGroupNotFound = 2

	codeGroupGidExists = 4

	codeGroupNameExists = 9
)

type groupClient struct {
	s system.System
}

func (c *groupClient) Get(ctx context.Context, gid int) (*Group, error) {
	cmd := NewCommand(fmt.Sprintf(`getent group %[1]d`, gid))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, errors.Join(ErrGroup, err)
	}

	if res.ExitCode == codeGroupNotFound {
		return nil, ErrGroupNotFound
	}
	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return nil, ErrGroupUnexpected
	}

	parsedGroup, err := parseGroupEntry(res.Stdout)
	if err != nil {
		return nil, ErrGroupUnexpected
	}

	groupSystem := parsedGroup.Gid < 1000

	group := &Group{
		Gid:    parsedGroup.Gid,
		Name:   parsedGroup.Name,
		System: groupSystem,
	}

	return group, nil
}

func (c *groupClient) Create(ctx context.Context, g Group) (int, error) {
	var args []string

	if g.Gid != -1 {
		args = append(args, fmt.Sprintf("--gid %d", g.Gid))
	}

	if g.System {
		args = append(args, "--system")
	}

	args = append(args, g.Name)

	cmd := NewCommand(fmt.Sprintf(`groupadd %[1]s && getent group %[2]s`, strings.Join(args, " "), g.Name))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return -1, errors.Join(ErrGroup, err)
	}

	switch res.ExitCode {
	case codeGroupGidExists:
		// GID not unique
		return -1, ErrGroupGidExists
	case codeGroupNameExists:
		// Group name not unique
		return -1, ErrGroupNameExists
	}

	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return -1, ErrGroupUnexpected
	}

	createdGroup, err := parseGroupEntry(res.Stdout)
	if err != nil {
		return -1, ErrGroupUnexpected
	}

	return createdGroup.Gid, nil
}

func (c *groupClient) Update(ctx context.Context, g Group) error {
	var args []string

	if g.Name != "" {
		args = append(args, fmt.Sprintf(`--new-name '%s'`, g.Name))
	}

	if len(args) == 0 {
		// Nothing to do because up-to-date
		return nil
	}

	groupmodCmd := fmt.Sprintf(`groupmod %s "${group}"`, strings.Join(args, " "))
	cmd := NewCommand(fmt.Sprintf(`_do() { gid=$1; group=$(getent group $gid | cut -d: -f1); [ ! -z "${group}" ] || return %[2]d; %[3]s; return $?; }; _do '%[1]d';`, g.Gid, codeGroupNotFound, groupmodCmd))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return errors.Join(ErrGroup, err)
	}

	switch res.ExitCode {
	case 0:
		// Success
		break
	case codeGroupNotFound:
		// Group not found
		return ErrGroupNotFound
	case codeGroupNameExists:
		// Group name not unique
		return ErrGroupNameExists
	default:
		return ErrGroupUnexpected
	}

	return nil
}

func (c *groupClient) Delete(ctx context.Context, gid int) error {
	cmd := NewCommand(fmt.Sprintf(`_do() { gid=$1; group=$(getent group $gid | cut -d: -f1); [ ! -z "${group}" ] || return %[2]d; groupdel "${group}"; return $?; }; _do '%[1]d';`, gid, codeGroupNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return errors.Join(ErrGroup, err)
	}

	// https://github.com/shadow-maint/shadow/blob/dc9fc048de56aa7b6eaf80b1c068a8b5d59b1bf0/src/groupdel.c#L77
	switch res.ExitCode {
	case 0:
		// Success
		break
	case codeGroupNotFound:
		// Not interpreted as error because this is the desired state
		break
	default:
		return errors.Join(ErrGroup, fmt.Errorf("failed to delete group with gid %d", gid))
	}

	return nil
}

type groupEntry struct {
	Name string
	Gid  int
}

func parseGroupEntry(data []byte) (*groupEntry, error) {
	parts := strings.Split(strings.TrimSpace(string(data)), ":")
	if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
		return nil, ErrGroupUnexpected
	}

	groupGid, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, ErrGroupUnexpected
	}

	return &groupEntry{
		Name: parts[0],
		Gid:  groupGid,
	}, nil
}
