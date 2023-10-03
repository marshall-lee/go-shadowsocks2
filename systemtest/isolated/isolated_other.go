//go:build !linux

package isolated

import (
	"net"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type Node struct{}

type Isolated struct {
	Client, Server, Peer Node
}

func New(t *testing.T) *Isolated {
	t.Skip("Isolation is not supported on", runtime.GOOS)
	return nil
}

func (_ Node) BanSourceIP(t *testing.T, ip net.IP) {
	require.Fail(t, "Isolation is not supported on", runtime.GOOS)
}

func (_ Node) SetupTCPRedirect(t *testing.T, redirPort, redir6Port uint16, mark uint32) {
	require.Fail(t, "Isolation is not supported on", runtime.GOOS)
}

func (_ Node) IPv4() net.IP {
	return nil
}

func (_ Node) IPv6() net.IP {
	return nil
}

func (_ Node) Do(t require.TestingT, f func()) {
	require.Fail(t, "Isolation is not supported on", runtime.GOOS)
}
