package systemtest

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func udpEcho(t *testing.T, ctx context.Context) (port uint16) {
	t.Helper()
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, "udp", "")
	require.NoError(t, err, "Failed to start udp listener")
	addr := conn.LocalAddr()
	_, portStr, err := net.SplitHostPort(addr.String())
	require.NoErrorf(t, err, "Failed to get udp listener port")
	port = parsePort(t, portStr)
	go func() {
		var buf [64 * 1024]byte
		for {
			n, addr, err := conn.ReadFrom(buf[:])
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if !assert.NoErrorf(t, err, "Failed to read from udp socket on %s", addr) {
				return
			}
			_, err = conn.WriteTo([]byte("ECHO "+string(buf[:n])), addr)
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if !assert.NoErrorf(t, err, "Failed to write to udp socket on %s", addr) {
				return
			}
		}
	}()
	t.Cleanup(func() {
		err := conn.Close()
		assert.NoErrorf(t, err, "Failed to close udp listener on %s", addr)
	})
	return
}
