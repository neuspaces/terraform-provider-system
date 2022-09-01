package sshclient

import (
	"context"
	"github.com/sethvargo/go-retry"
	"golang.org/x/crypto/ssh"
	"net"
)

type ConnectFunc func(ctx context.Context) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error)

type ConnectMiddleware func(ConnectFunc) ConnectFunc

// Prepare returns a ConnectFunc and validates the provided ConnectOption immediately
func Prepare(opts ...ConnectOption) (ConnectFunc, error) {
	args, err := processConnectOpts(opts)
	if err != nil {
		return nil, err
	}
	return ConnectCustom(args.netConnectFunc, args.addr, args.clientConfig), nil
}

// Connect returns a ConnectFunc which validates the provided ConnectOption when the connection is established.
func Connect(opts ...ConnectOption) ConnectFunc {
	return func(ctx context.Context) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
		args, err := processConnectOpts(opts)
		if err != nil {
			return nil, nil, nil, err
		}

		connectFunc := ConnectCustom(args.netConnectFunc, args.addr, args.clientConfig)

		return connectFunc(ctx)
	}
}

// ConnectCustom returns a ConnectFunc to a remote net.Addr using a connection method NetConnectFunc.
// The underlying ssh client is configured in ssh.ClientConfig.
// ConnectCustom allows for the highest d be used when more convenient ConnectFunc factory functions cannot be applied
func ConnectCustom(netConnectFunc NetConnectFunc, addr net.Addr, clientConfig *ssh.ClientConfig) ConnectFunc {
	return func(ctx context.Context) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
		conn, err := netConnectFunc(ctx)
		if err != nil {
			if conn != nil {
				_ = conn.Close()
			}
			return nil, nil, nil, err
		}

		return ssh.NewClientConn(conn, addr.String(), clientConfig)
	}
}

// CircuitBreak wraps a ConnectFunc and ensures that the ConnectFunc will not be called again when a previous
// call returned an error. If the ConnectFunc has returned an error previously, CircuitBreak will not call it
// again and return the same error.
func CircuitBreak() ConnectMiddleware {
	return func(next ConnectFunc) ConnectFunc {
		var connectErr error

		return func(ctx context.Context) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
			if connectErr != nil {
				return nil, nil, nil, connectErr
			}

			conn, chans, reqs, err := next(ctx)

			if err != nil {
				connectErr = err
				return nil, nil, nil, err
			}

			return conn, chans, reqs, nil
		}
	}
}

// Retry wraps a NetConnectFunc and attempts retries according to the provided retry.Backoff
func Retry(backoff retry.Backoff) ConnectMiddleware {
	return func(next ConnectFunc) ConnectFunc {
		return func(ctx context.Context) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
			var conn ssh.Conn
			var chans <-chan ssh.NewChannel
			var reqs <-chan *ssh.Request

			retryErr := retry.Do(ctx, backoff, func(ctx context.Context) error {
				var err error

				conn, chans, reqs, err = next(ctx)

				if err != nil {
					switch err.(type) {
					case *net.OpError:
						return retry.RetryableError(err)
					}

					switch err.Error() {
					case "ssh: handshake failed: EOF":
						return retry.RetryableError(err)
					}

					return err
				}
				return nil
			})
			if retryErr != nil {
				return nil, nil, nil, retryErr
			}

			return conn, chans, reqs, nil
		}
	}
}
