package systemtest

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tcpEcho(t *testing.T, ctx context.Context) (port uint16) {
	t.Helper()
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", "")
	require.NoError(t, err, "Failed to start tcp listener")
	addr := listener.Addr()
	_, portStr, err := net.SplitHostPort(addr.String())
	require.NoError(t, err, "Failed to get tcp listener port")
	port = parsePort(t, portStr)

	go func() {
		for {
			conn, err := listener.Accept()
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if !assert.NoErrorf(t, err, "Failed to start accept tcp connection on %s", addr) {
				return
			}
			go func() {
				defer func() { assert.NoError(t, conn.Close()) }()

				r := bufio.NewReader(conn)
				for {
					str, err := r.ReadString('\n')
					if err == io.EOF {
						return
					}
					if !assert.NoErrorf(t, err, "Failed to read from tcp connection on %s", addr) {
						return
					}
					_, err = conn.Write([]byte("ECHO " + str))
					if !assert.NoErrorf(t, err, "Failed to write to tcp connection on %s", addr) {
						return
					}
				}
			}()
		}
	}()
	t.Cleanup(func() {
		err := listener.Close()
		assert.NoErrorf(t, err, "Failed to close tcp listener on %s", addr)
	})
	return
}
