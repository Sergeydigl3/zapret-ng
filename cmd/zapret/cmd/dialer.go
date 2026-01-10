package cmd

import (
	"context"
	"net"
)

// UnixDialer creates a dialer function for Unix sockets.
func UnixDialer(socketPath string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", socketPath)
	}
}
