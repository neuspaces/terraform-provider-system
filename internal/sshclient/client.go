package sshclient

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
)

// Client is high level ssh client which relies on golang.org/x/crypto/ssh
type Client struct {
	*ssh.Client

	connectFunc ConnectFunc
}

var _ io.Closer = &Client{}

func New(connectFunc ConnectFunc) *Client {
	c := &Client{
		connectFunc: connectFunc,
	}

	return c
}

// Connected ensures that an ssh connection is established.
func (c *Client) Connected(ctx context.Context) error {
	if c.Client == nil {
		return c.Connect(ctx)
	}

	return nil
}

// Connect establishes an ssh connection.
func (c *Client) Connect(ctx context.Context) error {
	var err error

	// Require connect function
	if c.connectFunc == nil {
		return fmt.Errorf("sshclient: improperly configured, expected connect function")
	}

	// Reset client
	if c.Client != nil {
		_ = c.Close()
		c.Client = nil
	}

	// Connect
	sshConn, chans, reqs, err := c.connectFunc(ctx)
	if err != nil {
		return err
	}

	// Create ssh client
	c.Client = ssh.NewClient(sshConn, chans, reqs)

	return nil
}

func (c *Client) Close() error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
