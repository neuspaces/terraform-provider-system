package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/stat"
	"github.com/neuspaces/terraform-provider-system/internal/lib/typederror"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io"
	"io/fs"
	"strconv"
	"strings"
)

type File struct {
	Path  string
	Mode  fs.FileMode
	User  string
	Uid   int
	Group string
	Gid   int

	// Content optionally contains the file contents when enabled with FileClientIncludeContent
	Content io.Reader
	Md5Sum  string
}

func newFileFromStat(s *stat.Stat) *File {
	return &File{
		Path:  s.Name,
		Mode:  s.Mode.ToFsFileMode(),
		User:  s.User,
		Uid:   s.Uid,
		Group: s.Group,
		Gid:   s.Gid,
	}
}

type FileClient interface {
	Get(ctx context.Context, path string) (*File, error)
	Create(ctx context.Context, f File) error
	Update(ctx context.Context, f File) error
	Delete(ctx context.Context, path string) error
}

type FileClientOpt func(c *fileClient)

func FileClientCompression(enabled bool) FileClientOpt {
	return func(c *fileClient) {
		c.compress = enabled
	}
}

func FileClientIncludeContent(include bool) FileClientOpt {
	return func(c *fileClient) {
		c.includeContent = include
	}
}

func NewFileClient(s system.System, opts ...FileClientOpt) FileClient {
	fc := &fileClient{
		s: s,
	}

	for _, opt := range opts {
		opt(fc)
	}

	return fc
}

var (
	ErrFile = typederror.NewRoot("file resource")

	ErrFileExists = typederror.New("file exists", ErrFile)

	ErrFileNotFound = typederror.New("file not found", ErrFile)

	ErrFileUnexpected = typederror.New("unexpected error", ErrFile)
)

const (
	codeFileUnexpected = 1

	codeFilePathExists = 16

	codeFileNotFound = 17
)

type fileClient struct {
	s system.System

	compress       bool
	includeContent bool
}

func (c *fileClient) Get(ctx context.Context, path string) (*File, error) {
	cmd := NewCommand(fmt.Sprintf(`_do() { path='%[1]s'; [ -f "${path}" ] || return %[2]d; { stat -c '%[3]s' "${path}" && md5sum "${path}"; } || return 1; }; _do;`, path, codeFileNotFound, stat.FormatJsonGnu))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, ErrFileUnexpected.Raise(err)
	}

	switch res.ExitCode {
	case codeFileNotFound:
		return nil, ErrFileNotFound
	}

	stdoutLines := strings.Split(strings.TrimSpace(string(res.Stdout)), "\n")

	if res.ExitCode != 0 || len(res.Stdout) == 0 || len(stdoutLines) != 2 {
		return nil, ErrFileUnexpected
	}

	statOut := []byte(stdoutLines[0])
	parsedStat, err := stat.ParseJsonFormat(statOut)
	if err != nil {
		return nil, ErrFileUnexpected.Raise(err)
	}

	file := newFileFromStat(parsedStat)

	md5Parts := strings.Split(stdoutLines[1], "  ")
	if len(md5Parts) != 2 {
		return nil, ErrFileUnexpected
	}

	// Convert md5 sum from hex to base64 encoding
	md5Hex, err := hex.DecodeString(md5Parts[0])
	if err != nil {
		return nil, ErrFileUnexpected.Raise(err)
	}

	file.Md5Sum = base64.StdEncoding.EncodeToString(md5Hex)

	// Get content if requested
	if c.includeContent {
		catCmd := NewCommand(fmt.Sprintf(`cat '%s'`, path))
		catRes, err := ExecuteCommand(ctx, c.s, catCmd)
		if err != nil {
			return nil, ErrFileUnexpected.Raise(err)
		}

		if err := catRes.Error(); err != nil {
			return nil, ErrFileUnexpected.Raise(err)
		}

		file.Content = bytes.NewReader(catRes.Stdout)
	}

	return file, nil
}

func (c *fileClient) Create(ctx context.Context, f File) error {
	pathSub := `"${path}"`

	var createCmds []Command
	var createCmdIn io.Reader

	if f.Content != nil {
		// File content is provided from io.Reader

		if !c.compress {
			// Without transport compression
			createCmds = append(createCmds, NewCommand(fmt.Sprintf(`cat - > %s`, pathSub)))
			createCmdIn = f.Content
		} else {
			// With transport compression
			createCmds = append(createCmds, NewCommand(fmt.Sprintf(`gzip -d > %s`, pathSub)))

			// Setup pipe to compress source
			pipeReader, pipeWriter := io.Pipe()
			gzipWriter, err := gzip.NewWriterLevel(pipeWriter, gzip.BestCompression)
			if err != nil {
				return ErrFile.Raise(err)
			}

			go func() {
				_, _ = io.Copy(gzipWriter, f.Content)
				_ = gzipWriter.Close()
				_ = pipeWriter.Close()
			}()

			// Remote command stdin is the pipe output
			createCmdIn = pipeReader
		}
	} else {
		// Create file without content
		createCmds = append(createCmds, NewCommand(fmt.Sprintf(`touch %s`, pathSub)))
	}

	if f.Mode != 0 {
		createCmds = append(createCmds, &ChmodCommand{Path: pathSub, Mode: f.Mode})
	}

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

	cmd := NewInputCommand(fmt.Sprintf(`_do() { path=$1; [ ! -e "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, f.Path, codeFilePathExists, CompositeCommand(createCmds).Command()), createCmdIn)
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFile.Raise(err)
	}

	switch res.ExitCode {
	case codeFilePathExists:
		return ErrFileExists
	}

	err = res.Error()
	if err != nil {
		return ErrFile.Raise(err)
	}

	return nil
}

func (c *fileClient) Update(ctx context.Context, f File) error {
	pathSub := `"${path}"`

	var updateCmds []Command
	var updateCmdIn io.Reader

	if f.Content != nil {
		// File content is provided from io.Reader

		if !c.compress {
			// Without transport compression
			updateCmds = append(updateCmds, NewCommand(fmt.Sprintf(`cat - > %s`, pathSub)))
			updateCmdIn = f.Content
		} else {
			// With transport compression
			updateCmds = append(updateCmds, NewCommand(fmt.Sprintf(`gzip -d > %s`, pathSub)))

			// Setup pipe to compress source
			pipeReader, pipeWriter := io.Pipe()
			gzipWriter, err := gzip.NewWriterLevel(pipeWriter, gzip.BestCompression)
			if err != nil {
				return ErrFile.Raise(err)
			}

			go func() {
				_, _ = io.Copy(gzipWriter, f.Content)
				_ = gzipWriter.Close()
				_ = pipeWriter.Close()
			}()

			// Remote command stdin is the pipe output
			updateCmdIn = pipeReader
		}
	}

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

	cmd := NewInputCommand(fmt.Sprintf(`_do() { path=$1; [ -f "${path}" ] || return %[2]d; { %[3]s; } || return 1; }; _do '%[1]s';`, f.Path, codeFileNotFound, CompositeCommand(updateCmds).Command()), updateCmdIn)
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFile.Raise(err)
	}

	switch res.ExitCode {
	case codeFileNotFound:
		return ErrFileNotFound
	}

	err = res.Error()
	if err != nil {
		return ErrFile.Raise(err)
	}

	return nil
}

func (c *fileClient) Delete(ctx context.Context, path string) error {
	cmd := NewCommand(fmt.Sprintf(`_do() { path=$1; [ -f "${path}" ] || return %[2]d; rm -f "${path}" || return 1; }; _do '%[1]s';`, path, codeFileNotFound))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return ErrFile.Raise(err)
	}

	switch res.ExitCode {
	case codeFileNotFound:
		return ErrFileNotFound
	}

	if res.ExitCode != 0 {
		return ErrFile.Raise(fmt.Errorf("failed to delete %q", path))
	}

	return nil
}
