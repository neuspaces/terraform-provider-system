package client

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io/fs"
	"strings"
)

type SimpleCommand string

var _ Command = SimpleCommand("")

func (a SimpleCommand) Command() string {
	return string(a)
}

type ChownCommand struct {
	Path  string
	User  string
	Group string

	// NoDereference: affect each symbolic link instead of any referenced file
	NoDereference bool
}

var _ Command = &ChownCommand{}

func (c *ChownCommand) Command() string {
	if c.Path == "" || c.User == "" {
		return ""
	}

	var args []string

	if c.NoDereference {
		args = append(args, `-h`)
	}

	if c.Group == "" {
		args = append(args, c.User)
	} else {
		args = append(args, fmt.Sprintf(`%s:%s`, c.User, c.Group))
	}

	args = append(args, c.Path)

	return fmt.Sprintf(`chown %s`, strings.Join(args, ` `))
}

type ChgrpCommand struct {
	Path  string
	Group string

	// NoDereference: affect each symbolic link instead of any referenced file
	NoDereference bool
}

var _ Command = &ChgrpCommand{}

func (c *ChgrpCommand) Command() string {
	if c.Path == "" || c.Group == "" {
		return ""
	}

	var args []string

	if c.NoDereference {
		args = append(args, `-h`)
	}

	args = append(args, c.Group, c.Path)

	return fmt.Sprintf(`chgrp %s`, strings.Join(args, ` `))
}

type ChmodCommand struct {
	Path string
	Mode fs.FileMode
}

var _ Command = &ChmodCommand{}

func (c *ChmodCommand) Command() string {
	mode := c.Mode & fs.ModePerm
	if c.Path == "" || mode == 0 {
		return ""
	}

	return fmt.Sprintf(`chmod %o %s`, mode, c.Path)
}

type MkdirCommand struct {
	Path string
	Mode fs.FileMode
}

var _ Command = &MkdirCommand{}

func (c *MkdirCommand) Command() string {
	if c.Path == "" {
		return ""
	}

	mode := c.Mode & fs.ModePerm
	if mode > 0 {
		return fmt.Sprintf(`mkdir -m %o -p %s`, mode, c.Path)
	}

	return fmt.Sprintf(`mkdir -p %s`, c.Path)
}

type CatCommand struct {
	Path string
}

var _ Command = &CatCommand{}

func (c *CatCommand) Command() string {
	if c.Path == "" {
		return ""
	}

	return fmt.Sprintf(`cat '%s'`, c.Path)
}

func cat(ctx context.Context, s system.System, path string) ([]byte, error) {
	catCmd := &CatCommand{Path: path}
	res, err := ExecuteCommand(ctx, s, catCmd)
	if err != nil {
		return nil, err
	}

	if res.ExitCode != 0 {
		return nil, fmt.Errorf("non-zero exit code: %d", res.ExitCode)
	}

	return res.Stdout, nil
}

type CompositeCommand []Command

var _ Command = CompositeCommand{}

func (c CompositeCommand) Command() string {
	var cmds []string
	for _, cmd := range c {
		cmds = append(cmds, cmd.Command())
	}
	return strings.Join(cmds, " && ")
}
