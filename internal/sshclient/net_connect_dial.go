package sshclient

import (
	"context"
	"net"
	"time"
)

// Dial returns a NetConnectFunc which connects to the remote via net.Dial
func Dial(addr net.Addr, timeout time.Duration) NetConnectFunc {
	return func(ctx context.Context) (net.Conn, error) {
		c, err := net.DialTimeout(addr.Network(), addr.String(), timeout)
		if err != nil {
			return nil, err
		}

		if tcpConn, ok := c.(*net.TCPConn); ok {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				return nil, err
			}
		}

		return c, nil
	}
}
