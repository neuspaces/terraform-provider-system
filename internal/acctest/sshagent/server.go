package sshagent

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
)

type Server struct {
	agent agent.Agent

	socketFileFunc func() (string, error)
	socketFile     string

	errCallback func(err error)
}

type ServerOption func(*Server) error

func SocketFile(socketFileName string) ServerOption {
	return func(s *Server) error {
		if _, err := os.Stat(socketFileName); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("sshagent.SocketFile: socket file %q already exists", socketFileName)
		}

		s.socketFile = socketFileName

		return nil
	}
}

func tempSocketFileFunc() (string, error) {
	// Random file within temp directory
	socketFile, err := ioutil.TempFile("", "*")
	if err != nil {
		return "", err
	}
	socketFileName := socketFile.Name()

	// Remove the temp file so that a socket file can be created with the same name
	err = os.Remove(socketFileName)
	if err != nil {
		return "", err
	}

	return socketFileName, nil
}

func TempSocketFile() ServerOption {
	return func(s *Server) error {
		socketFileName, err := tempSocketFileFunc()
		if err != nil {
			return nil
		}
		s.socketFile = socketFileName

		return nil
	}
}

// PrivateKey is an option which adds the provided private key to the agent.
// PrivateKey builds on agent.AddedKey
// privateKey must be either a *rsa.PrivateKey, *dsa.PrivateKey, *ed25519.PrivateKey or *ecdsa.PrivateKey, or if a string or []byte will be parsed before adding.
func PrivateKey(privateKey interface{}) ServerOption {
	return func(s *Server) error {
		switch privateKeyVal := privateKey.(type) {
		case string:
			parsedPrivateKey, err := ssh.ParseRawPrivateKey([]byte(privateKeyVal))
			if err != nil {
				return nil
			}
			privateKey = parsedPrivateKey
		case []byte:
			parsedPrivateKey, err := ssh.ParseRawPrivateKey(privateKeyVal)
			if err != nil {
				return nil
			}
			privateKey = parsedPrivateKey
		}

		err := s.agent.Add(agent.AddedKey{
			PrivateKey: privateKey,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func NewServer(opts ...ServerOption) (*Server, error) {
	s := &Server{
		agent: agent.NewKeyring(),
		// Default error callback does nothing
		errCallback: func(err error) {},
		// By default create a socket file in temp directory
		socketFileFunc: tempSocketFileFunc,
	}

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) SocketFile() string {
	return s.socketFile
}

func (s *Server) Agent() agent.Agent {
	return s.agent
}

// ListenAndServe listens on the socket file and then calls Serve to handle requests on incoming connections.
// ListenAndServe blocks until the provided context.Context is cancelled
func (s *Server) ListenAndServe(ctx context.Context) error {
	listener, err := s.Listen()
	if err != nil {
		return err
	}

	return s.Serve(ctx, listener)
}

// Listen creates a listener on the socket file
// After Listen, the socket file name is available via SocketFile
func (s *Server) Listen() (net.Listener, error) {
	if s.socketFile == "" {
		if s.socketFileFunc == nil {
			return nil, fmt.Errorf("sshagent.Serve: socket path not defined")
		}

		socketFile, err := s.socketFileFunc()
		if err != nil {
			return nil, fmt.Errorf("sshagent.Serve: failed to create socket path: %w", err)
		}

		s.socketFile = socketFile
	}

	listener, err := net.Listen("unix", s.socketFile)
	if err != nil {
		return nil, err
	}

	return listener, nil
}

// Serve accepts incoming connections on the Listener l, creating a new goroutine for each.
// ListenAndServe blocks until the provided context.Context is cancelled
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	// Wait group to keep track of active session
	var sessions sync.WaitGroup

	// Main loop
	go func() {
		for {
			// Accept new connection; blocks until new connection
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					s.errCallback(fmt.Errorf("sshagent.Serve: error accepting connection: %v", err))
					continue
				}
			}

			sessions.Add(1)

			go func() {
				// agent.ServeAgent always returns an error
				// io.EOF is returned when successful
				err := agent.ServeAgent(s.agent, conn)
				if err != io.EOF {
					s.errCallback(fmt.Errorf("sshagent.Serve: error serving agent: %v", err))
				}

				sessions.Done()
			}()
		}
	}()

	// Wait for context
	<-ctx.Done()

	// Collect error which occur during shutdown
	var shutdownErrs error

	// Shutdown listener
	if listener != nil {
		err := listener.Close()
		if err != nil {
			shutdownErrs = multierror.Append(shutdownErrs, err)
		}
	}

	// Wait for sessions to complete
	sessions.Wait()

	if shutdownErrs != nil {
		return fmt.Errorf("sshagent.Serve: one or more errors occurred during shutdown: %v", shutdownErrs.Error())
	}

	return nil
}
