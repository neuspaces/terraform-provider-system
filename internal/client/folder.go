package client

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/stat"
	"github.com/neuspaces/terraform-provider-system/internal/lib/typederror"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io/fs"
	"strconv"
)

type Folder struct {
	Path  string
	Mode  fs.FileMode
	User  string
	Uid   int
	Group string
	Gid   int
}

func newFolderFromStat(s *stat.Stat) *Folder {
	return &Folder{
		Path:  s.Name,
		Mode:  s.Mode.ToFsFileMode(),
		User:  s.User,
		Uid:   s.Uid,
		Group: s.Group,
		Gid:   s.Gid,
	}
}

type FolderClient interface {
	Get(ctx context.Context, path string) (*Folder, error)
	Create(ctx context.Context, folder Folder) error
	Update(ctx context.Context, folder Folder) error
	Delete(ctx context.Context, path string) error
}

func NewFolderClient(s system.System) FolderClient {
	return &folderClient{
		s: s,
	}
}

var (
	ErrFolder = typederror.NewRoot("folder resource")

	ErrFolderPathExists = typederror.New("folder path exists", ErrFile)

	ErrFolderNotFound = typederror.New("folder not found", ErrFile)

	ErrFolderUnexpected = typederror.New("unexpected error", ErrFile)
)

const (
	codeFolderUnexpected = 1

	codeFolderPathExists = 16

	codeFolderNotFound = 17
)

type folderClient struct {
	s system.System
}

func (c *folderClient) Get(ctx context.Context, path string) (*Folder, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -d "${path}" ] || return %[2]d; stat -c '%[3]s' "${path}" || return 1; }; _do '%[1]s';`, path, codeFolderNotFound, stat.FormatJsonGnu))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, ErrFolder.Raise(err)
	}

	switch res.ExitCode {
	case codeFolderNotFound:
		return nil, ErrFolderNotFound
	}

	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return nil, ErrFolderUnexpected
	}

	parsedStat, err := stat.ParseJsonFormat(res.Stdout)
	if err != nil {
		return nil, ErrFolder.Raise(err)
	}

	folder := newFolderFromStat(parsedStat)

	return folder, nil
}

func (c *folderClient) Create(ctx context.Context, f Folder) error {
	pathSub := `"${path}"`

	var createCmds []Command

	createCmds = append(createCmds, &MkdirCommand{Path: pathSub, Mode: f.Mode})

	if f.Uid != -1 {
		createCmds = append(createCmds, &ChownCommand{Path: pathSub, User: strconv.Itoa(f.Uid)})
	} else if f.User != "" {
		createCmds = append(createCmds, &ChownCommand{Path: pathSub, User: f.User})
	}

	if f.Gid != -1 {
		createCmds = append(createCmds, &ChgrpCommand{Path: pathSub, Group: strconv.Itoa(f.Gid)})
	} else if f.Group != "" {
		createCmds = append(createCmds, &ChgrpCommand{Path: pathSub, Group: f.Group})
	}

	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ ! -e "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, f.Path, codeFolderPathExists, CompositeCommand(createCmds).Command()))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFolder.Raise(err)
	}

	switch res.ExitCode {
	case codeFolderPathExists:
		return ErrFolderPathExists
	}

	err = res.Error()
	if err != nil {
		return ErrFolder.Raise(err)
	}

	return nil
}

func (c *folderClient) Update(ctx context.Context, f Folder) error {
	pathSub := `"${path}"`

	var updateCmds []Command

	if f.Mode != 0 {
		updateCmds = append(updateCmds, &ChmodCommand{Path: pathSub, Mode: f.Mode})
	}

	if f.Uid != -1 {
		updateCmds = append(updateCmds, &ChownCommand{Path: pathSub, User: strconv.Itoa(f.Uid)})
	} else if f.User != "" {
		updateCmds = append(updateCmds, &ChownCommand{Path: pathSub, User: f.User})
	}

	if f.Gid != -1 {
		updateCmds = append(updateCmds, &ChgrpCommand{Path: pathSub, Group: strconv.Itoa(f.Gid)})
	} else if f.Group != "" {
		updateCmds = append(updateCmds, &ChgrpCommand{Path: pathSub, Group: f.Group})
	}

	if len(updateCmds) == 0 {
		// Nothing to do because up-to-date
		return nil
	}

	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -d "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, f.Path, codeFolderNotFound, CompositeCommand(updateCmds).Command()))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFolder.Raise(err)
	}

	switch res.ExitCode {
	case codeFolderNotFound:
		return ErrFolderNotFound
	}

	err = res.Error()
	if err != nil {
		return ErrFolder.Raise(err)
	}

	return nil
}

func (c *folderClient) Delete(ctx context.Context, path string) error {
	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -d "${path}" ] || return %[2]d; rm -rf "${path}" || return 1; }; _do '%[1]s';`, path, codeFolderNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFolder.Raise(err)
	}

	switch res.ExitCode {
	case codeFolderNotFound:
		return ErrFolderNotFound
	}

	if res.ExitCode != 0 {
		return ErrFolder.Raise(fmt.Errorf("failed to delete %q", path))
	}

	return nil
}
