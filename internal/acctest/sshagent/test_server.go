package sshagent

import (
	"context"
	"os"
	"sync"
	"testing"
)

// testServerSshAuthSockEnvM locks the use of the environment variable SSH_AUTH_SOCK
var testServerSshAuthSockEnvM sync.Mutex

const EnvVarSshAuthSock = "SSH_AUTH_SOCK"

// TestServer helps to run a Server within tests
type TestServer struct {
	server *Server
}

// NewTestServer returns a TestServer
func NewTestServer(t *testing.T, opts ...ServerOption) *TestServer {
	s, err := NewServer(opts...)
	if err != nil {
		t.Fatal(err)
	}

	return &TestServer{
		server: s,
	}
}

// Use ensure that the underlying Server is started before scopeFunc and shutdown afterwards.
// Use sets the environment variable SSH_AUTH_SOCK within the scopeFunc
func (s *TestServer) Use(t *testing.T, scopeFunc func(t *testing.T)) {
	// Prepare context
	ctx, agentShutdown := context.WithCancel(context.Background())
	defer agentShutdown()

	// Start listening
	listener, err := s.server.Listen()
	if err != nil {
		t.Fatal(err)
	}

	// Start serving; the result is collected after the scope has run because testing.T must not be used from another goroutine
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.server.Serve(ctx, listener)
	}()

	// Mutex on SSH_AUTH_SOCK; only a single thread may set SSH_AUTH_SOCK
	testServerSshAuthSockEnvM.Lock()
	defer testServerSshAuthSockEnvM.Unlock()

	// Get current SSH_AUTH_SOCK
	prevSshAuthSock, hasPrevSshAuthSock := os.LookupEnv(EnvVarSshAuthSock)

	// Set SSH_AUTH_SOCK with scope
	_ = os.Setenv(EnvVarSshAuthSock, s.server.SocketFile())

	// Run scoped func
	scopeFunc(t)

	// Stop agent server and surface shutdown errors in the test goroutine
	agentShutdown()
	if err := <-serveErr; err != nil {
		t.Error(err)
	}

	// Pop SSH_AUTH_SOCK
	if hasPrevSshAuthSock {
		_ = os.Setenv(EnvVarSshAuthSock, prevSshAuthSock)
	} else {
		_ = os.Unsetenv(EnvVarSshAuthSock)
	}
}
