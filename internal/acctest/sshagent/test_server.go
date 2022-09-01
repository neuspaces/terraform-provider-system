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
	var err error

	// Prepare context
	ctx, agentShutdown := context.WithCancel(context.Background())
	defer func() {
		agentShutdown()
	}()

	// Start listening
	listener, err := s.server.Listen()
	if err != nil {
		t.Fatal(err)
	}

	// Start serving
	go func() {
		err = s.server.Serve(ctx, listener)
		if err != nil {
			t.Fatal(err)
		}
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

	// Stop agent server
	agentShutdown()

	// Pop SSH_AUTH_SOCK
	if hasPrevSshAuthSock {
		_ = os.Setenv(EnvVarSshAuthSock, prevSshAuthSock)
	}
}
