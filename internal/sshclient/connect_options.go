package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type ConnectOption func(*connectArgs) error

type connectArgs struct {
	netConnectFunc NetConnectFunc
	addr           net.Addr
	clientConfig   *ssh.ClientConfig
}

func processConnectOpts(opts []ConnectOption) (*connectArgs, error) {
	args := &connectArgs{
		clientConfig: &ssh.ClientConfig{
			HostKeyCallback: unconfiguredHostKey(),
		},
	}

	for _, opt := range opts {
		err := opt(args)
		if err != nil {
			return nil, err
		}
	}

	return args, nil
}

func Addr(addr net.Addr) ConnectOption {
	return func(c *connectArgs) error {
		c.addr = addr
		return nil
	}
}

func Net(netConnectFunc NetConnectFunc) ConnectOption {
	return func(c *connectArgs) error {
		c.netConnectFunc = netConnectFunc
		return nil
	}
}

func User(user string) ConnectOption {
	return func(c *connectArgs) error {
		c.clientConfig.User = user
		return nil
	}
}

func Timeout(timeout time.Duration) ConnectOption {
	return func(c *connectArgs) error {
		c.clientConfig.Timeout = timeout
		return nil
	}
}
func HostKey(hostKeyVerifier HostKeyVerifier) ConnectOption {
	return func(c *connectArgs) error {
		callback, err := hostKeyVerifier()
		if err != nil {
			return err
		}
		c.clientConfig.HostKeyCallback = callback
		return nil
	}
}

func HostKeyCallback(callback ssh.HostKeyCallback) ConnectOption {
	return func(c *connectArgs) error {
		c.clientConfig.HostKeyCallback = callback
		return nil
	}
}

func Auth(authMethod AuthMethod) ConnectOption {
	return func(c *connectArgs) error {
		if authMethod == nil {
			return fmt.Errorf("sshclient: expected auth method, got nil")
		}

		am, err := authMethod()
		if err != nil {
			return err
		}

		for _, a := range am {
			c.clientConfig.Auth = append(c.clientConfig.Auth, a)
		}
		return nil
	}
}
