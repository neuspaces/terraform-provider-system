package sshclient

import (
	"context"
	"net"
)

// Proxy returns a NetConnectFunc which connects to the remote via a proxy Client
func Proxy(proxy *Client, addr net.Addr) NetConnectFunc {
	return func(ctx context.Context) (net.Conn, error) {
		// Connect to proxy
		err := proxy.Connect(ctx)
		if err != nil {
			return nil, err
		}

		// Connect to remote via proxy
		conn, err := proxy.Dial(addr.Network(), addr.String())
		if err != nil {
			if err := proxy.Close(); err != nil {
				return nil, err
			}

			return nil, err
		}

		return &proxyConn{
			Conn:  conn,
			Proxy: proxy,
		}, nil
	}
}

type proxyConn struct {
	net.Conn
	Proxy *Client
}

func (c *proxyConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return err
	}
	return c.Proxy.Close()
}
