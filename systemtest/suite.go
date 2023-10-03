package systemtest

import (
	"bytes"
	"strings"
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/systemtest/child"
	"github.com/shadowsocks/go-shadowsocks2/systemtest/isolated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/proxy"
)

var shadowsocksPath string

const (
	shadowsocksPort = 8488
	testTimeout     = 10 * time.Second // Timeout of a single test
	networkTimeout  = time.Second      // Short timeout for network operations
	eventuallyTick  = 50 * time.Millisecond
)

type SystemTestSuite struct {
	suite.Suite
	testDeadline           time.Time
	ctx                    context.Context
	isolated               *isolated.Isolated
	clientIPv4, clientIPv6 net.IP
	serverIPv4, serverIPv6 net.IP
	peerIPv4, peerIPv6     net.IP
}

func (s *SystemTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	s.T().Cleanup(cancel)
	s.ctx = ctx
	testDeadline, ok := ctx.Deadline()
	s.Require().True(ok)
	s.testDeadline = testDeadline

	s.isolated = nil
	ipv4loopback := net.IPv4(127, 0, 0, 1)
	s.clientIPv4 = ipv4loopback
	s.clientIPv6 = net.IPv6loopback
	s.serverIPv4 = ipv4loopback
	s.serverIPv6 = net.IPv6loopback
	s.peerIPv4 = ipv4loopback
	s.peerIPv6 = net.IPv6loopback
}

func (s *SystemTestSuite) setupIsolation() {
	t := s.T()
	if runtime.GOOS != "linux" {
		t.Skip("Isolation is not supported on", runtime.GOOS)
	}

	s.isolated = isolated.New(t)
	s.clientIPv4 = s.isolated.Client.IPv4()
	s.clientIPv6 = s.isolated.Client.IPv6()
	s.serverIPv4 = s.isolated.Server.IPv4()
	s.serverIPv6 = s.isolated.Server.IPv6()
	s.peerIPv4 = s.isolated.Peer.IPv4()
	s.peerIPv6 = s.isolated.Peer.IPv6()

	// Don't allow to directly connect from the client to peer.
	s.isolated.Peer.BanSourceIP(t, s.clientIPv4)
	s.isolated.Peer.BanSourceIP(s.T(), s.clientIPv6)
}

func (s *SystemTestSuite) serverHostPort() string {
	return joinIPPort(s.serverIPv4, shadowsocksPort)
}

func (s *SystemTestSuite) networkDeadline() time.Time {
	networkDeadline := time.Now().Add(networkTimeout)
	if s.testDeadline.Before(networkDeadline) {
		return s.testDeadline
	}
	return networkDeadline
}

func (s *SystemTestSuite) shadowsocks(args ...string) {
	t := s.T()
	t.Helper()
	output := new(bytes.Buffer)
	child.Start(t, output, shadowsocksPath, args...)
	t.Cleanup(func() {
		// Print the output for debugging purposes
		if t.Failed() {
			fmt.Fprintf(os.Stderr, "PROCESS OUTPUT %s %s:\n", shadowsocksPath, strings.Join(args, " "))
			fmt.Fprintln(os.Stderr, output.String())
		}
	})
}

func (s *SystemTestSuite) eventually(condition func(collect *assert.CollectT), msgAndArgs ...interface{}) {
	s.T().Helper()
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		if s.isolated != nil {
			s.isolated.Client.Do(c, func() { condition(c) })
		} else {
			condition(c)
		}
	}, time.Until(s.testDeadline), eventuallyTick, msgAndArgs...)
}

func (s *SystemTestSuite) dial(dialer proxy.ContextDialer, network, address string) (conn net.Conn, err error) {
	s.T().Helper()
	ctx, cancel := context.WithTimeout(s.ctx, networkTimeout)
	defer cancel()
	conn, err = dialer.DialContext(ctx, network, address)
	if err == nil {
		err = conn.SetDeadline(s.networkDeadline())

	}
	return
}

func joinIPPort(ip net.IP, port uint16) string {
	return net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
}

func parsePort(t require.TestingT, str string) uint16 {
	port, err := strconv.ParseUint(str, 10, 16)
	require.NoError(t, err, "Failed to parse a port string", str)
	return uint16(port)
}
