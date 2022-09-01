package ssh

import (
	"bytes"
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/lib/stat"
	"github.com/neuspaces/terraform-provider-system/internal/sshclient"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
	"io/fs"
	"sync"
)

type System struct {
	sshClient  *sshclient.Client
	sshClientM sync.Mutex

	cmdM cmd.Middleware

	sessions *semaphore.Weighted
}

// System implements system.System
var _ system.System = &System{}

type SystemOption func(*System) error

// Sessions is a SystemOption which defines the maximum number of concurrent ssh sessions to the remote
// Set Sessions to 0 to not limit the number of concurrent ssh sessions.
func Sessions(c int) SystemOption {
	return func(s *System) error {
		if c > 0 {
			s.sessions = semaphore.NewWeighted(int64(c))
		} else if c == 0 {
			s.sessions = nil
		} else {
			return fmt.Errorf("invalid connections value: %d", c)
		}

		return nil
	}
}

func CommandMiddleware(m cmd.Middleware) SystemOption {
	return func(s *System) error {
		s.cmdM = m
		return nil
	}
}

func NewSystem(sshClient *sshclient.Client, opts ...SystemOption) (*System, error) {
	var err error

	s := &System{
		sshClient: sshClient,
	}

	// Apply options
	for _, opt := range opts {
		err = opt(s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Close closes the underlying ssh client
func (s *System) Close() error {
	s.sshClientM.Lock()
	defer s.sshClientM.Unlock()

	if s.sshClient != nil {
		return s.sshClient.Close()
	}

	return nil
}

// Open returns a fs.File
// Open fetches fs.FileInfo immediately when invoked
func (s *System) Open(ctx context.Context, name string) (fs.File, error) {
	// Retrieve file info
	fileInfo, err := s.Stat(ctx, name)
	if err != nil {
		return nil, err
	}

	// Return readFile with file info and reader based on `cat` command
	file := &readFile{
		fileInfo:   fileInfo,
		readCloser: newCatFileReader(ctx, s, name),
	}

	return file, nil
}

func (s *System) Create(ctx context.Context, fileInfo fs.FileInfo) (system.WriteFile, error) {
	return nil, fmt.Errorf("create is not supported by ssh.System")
}

func (s *System) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	statOut := &bytes.Buffer{}
	statCmd := cmd.NewCommand(fmt.Sprintf(`stat -t '%s'`, name), cmd.Stdout(statOut))

	statCmdResult, err := s.Execute(ctx, statCmd)
	if err != nil {
		return nil, err
	}

	if statCmdResult.ExitCode() != 0 {
		return nil, fs.ErrNotExist
	}

	fileInfo, err := stat.ParseTerseFormat(statOut.Bytes())
	if err != nil {
		return nil, err
	}

	return fileInfo.ToFsFileInfo(), nil
}

func (s *System) Execute(ctx context.Context, c cmd.Command) (cmd.Result, error) {
	var err error

	if s.sessions != nil {
		err := s.sessions.Acquire(ctx, 1)
		if err != nil {
			return nil, err
		}
		defer s.sessions.Release(1)
	}

	// Apply command middleware
	if s.cmdM != nil {
		c = s.cmdM(c)
	}

	// Ensure connection
	s.sshClientM.Lock()
	err = s.sshClient.Connected(ctx)
	s.sshClientM.Unlock()
	if err != nil {
		return nil, err
	}

	// Create session
	sess, err := s.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close session
		_ = sess.Close()
	}()

	// Connect inputs/outputs
	sess.Stdin = c.Stdin()
	sess.Stdout = c.Stdout()
	sess.Stderr = c.Stderr()

	// Completed channel is closed after sess.Wait returned
	completed := make(chan struct{})

	// Start command
	err = sess.Start(c.Command())
	if err != nil {
		return nil, err
	}

	// Handle context cancellation
	var signalErr error
	go func() {
		select {
		case <-ctx.Done():
			// Parent context has been cancelled

			// Send interrupt (SIGINT) when context is cancelled
			err := sess.Signal(ssh.SIGINT)
			if err != nil {
				signalErr = fmt.Errorf("ssh.System: failed to send singal SIGINT: %w", err)
			}

			// Close session (will lead sess.Wait to return)
			_ = sess.Close()

			return
		case <-completed:
			// Command completed
			return
		}
	}()

	// Wait for command to complete
	err = sess.Wait()
	if signalErr != nil {
		return nil, signalErr
	} else if err != nil {
		if exitErr, isExitErr := err.(*ssh.ExitError); isExitErr {
			return cmd.NewResult(exitErr.ExitStatus()), nil
		} else {
			return nil, err
		}
	}

	close(completed)

	return cmd.NewResult(0), nil
}
