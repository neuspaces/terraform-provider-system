package client

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/stat"
	"github.com/neuspaces/terraform-provider-system/internal/lib/typederror"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"strconv"
)

type Link struct {
	Path   string
	Target string
	User   string
	Uid    int
	Group  string
	Gid    int
}

func newLinkFromStat(s *stat.Stat) *Link {
	return &Link{
		Path:   s.Name,
		Target: s.Target,
		User:   s.User,
		Uid:    s.Uid,
		Group:  s.Group,
		Gid:    s.Gid,
	}
}

type LinkClient interface {
	Get(ctx context.Context, path string) (*Link, error)
	Create(ctx context.Context, l Link) error
	Update(ctx context.Context, l Link) error
	Delete(ctx context.Context, path string) error
}

func NewLinkClient(s system.System) LinkClient {
	return &linkClient{
		s: s,
	}
}

var (
	ErrLink = typederror.NewRoot("link resource")

	ErrLinkExists = typederror.New("link exists", ErrLink)

	ErrLinkNotFound = typederror.New("link not found", ErrLink)

	ErrLinkUnexpected = typederror.New("unexpected error", ErrLink)
)

const (
	codeLinkUnexpected = 1

	codeLinkPathExists = 16

	codeLinkNotFound = 17
)

type linkClient struct {
	s system.System
}

func (c *linkClient) Get(ctx context.Context, path string) (*Link, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -L "${path}" ] || return %[2]d; stat -c '%[3]s' "${path}" || return 1; }; _do '%[1]s';`, path, codeLinkNotFound, stat.FormatJsonGnu))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, ErrLink.Raise(err)
	}

	switch res.ExitCode {
	case codeLinkNotFound:
		return nil, ErrLinkNotFound
	}

	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return nil, ErrLinkUnexpected
	}

	parsedStat, err := stat.ParseJsonFormat(res.Stdout)
	if err != nil {
		return nil, ErrLink.Raise(err)
	}

	link := newLinkFromStat(parsedStat)

	return link, nil
}

func (c *linkClient) Create(ctx context.Context, l Link) error {
	pathSub := `"${path}"`

	var createCmds []Command

	createCmds = append(createCmds, NewCommand(fmt.Sprintf(`ln -s '%s' %s`, l.Target, pathSub)))

	if l.Uid != -1 {
		createCmds = append(createCmds, &ChownCommand{Path: pathSub, User: strconv.Itoa(l.Uid), NoDereference: true})
	} else if l.User != "" {
		createCmds = append(createCmds, &ChownCommand{Path: pathSub, User: l.User, NoDereference: true})
	}

	if l.Gid != -1 {
		createCmds = append(createCmds, &ChgrpCommand{Path: pathSub, Group: strconv.Itoa(l.Gid), NoDereference: true})
	} else if l.Group != "" {
		createCmds = append(createCmds, &ChgrpCommand{Path: pathSub, Group: l.Group, NoDereference: true})
	}

	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ ! -e "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, l.Path, codeLinkPathExists, CompositeCommand(createCmds).Command()))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrLink.Raise(err)
	}

	switch res.ExitCode {
	case codeLinkPathExists:
		return ErrLinkExists
	}

	err = res.Error()
	if err != nil {
		return ErrLink.Raise(err)
	}

	return nil
}

func (c *linkClient) Update(ctx context.Context, l Link) error {
	pathSub := `"${path}"`

	var updateCmds []Command

	if l.Target != "" {
		updateCmds = append(updateCmds, NewCommand(fmt.Sprintf(`ln -sf '%s' %s`, l.Target, pathSub)))
	}

	if l.Uid != -1 {
		updateCmds = append(updateCmds, &ChownCommand{Path: pathSub, User: strconv.Itoa(l.Uid)})
	} else if l.User != "" {
		updateCmds = append(updateCmds, &ChownCommand{Path: pathSub, User: l.User})
	}

	if l.Gid != -1 {
		updateCmds = append(updateCmds, &ChgrpCommand{Path: pathSub, Group: strconv.Itoa(l.Gid)})
	} else if l.Group != "" {
		updateCmds = append(updateCmds, &ChgrpCommand{Path: pathSub, Group: l.Group})
	}

	if len(updateCmds) == 0 {
		// Nothing to do because up-to-date
		return nil
	}

	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -L "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, l.Path, codeLinkNotFound, CompositeCommand(updateCmds).Command()))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrLink.Raise(err)
	}

	switch res.ExitCode {
	case codeLinkNotFound:
		return ErrLinkNotFound
	}

	err = res.Error()
	if err != nil {
		return ErrLink.Raise(err)
	}

	return nil
}

func (c *linkClient) Delete(ctx context.Context, path string) error {
	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -L "${path}" ] || return %[2]d; rm -f "${path}" || return 1; }; _do '%[1]s';`, path, codeLinkNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrLink.Raise(err)
	}

	switch res.ExitCode {
	case codeLinkNotFound:
		return ErrLinkNotFound
	}

	if res.ExitCode != 0 {
		return ErrFile.Raise(fmt.Errorf("failed to delete %q", path))
	}

	return nil
}
