package sshclient

import (
	"context"
	"github.com/sethvargo/go-retry"
	"net"
)

type NetConnectFunc func(context.Context) (net.Conn, error)

type NetConnectMiddleware func(NetConnectFunc) NetConnectFunc

// NetCircuitBreaker wraps a NetConnectFunc and ensures that the NetConnectFunc will not be called again when a previous
// call returned an error. If the NetConnectFunc has returned an error previously, NetCircuitBreaker will not call it
// again and return the same error.
func NetCircuitBreaker() NetConnectMiddleware {
	return func(next NetConnectFunc) NetConnectFunc {
		var connectErr error

		return func(ctx context.Context) (net.Conn, error) {
			if connectErr != nil {
				return nil, connectErr
			}

			conn, err := next(ctx)

			if err != nil {
				connectErr = err
				return nil, err
			}

			return conn, nil
		}
	}
}

// NetRetry wraps a NetConnectFunc and attempts retries according to the provided retry.Backoff
func NetRetry(backoff retry.Backoff) NetConnectMiddleware {
	return func(next NetConnectFunc) NetConnectFunc {
		return func(ctx context.Context) (net.Conn, error) {
			var conn net.Conn

			retryErr := retry.Do(ctx, backoff, func(ctx context.Context) error {
				var err error
				conn, err = next(ctx)
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
				return nil, retryErr
			}

			return conn, nil
		}
	}
}
