package systemtest

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	redirFwmark = 0x1234
	redirPort   = 1082
	redir6Port  = 1083
)

type RedirTestSuite struct {
	SystemTestSuite
}

func TestRedir(t *testing.T) {
	suite.Run(t, new(RedirTestSuite))
}

func (s *RedirTestSuite) SetupTest() {
	s.SystemTestSuite.SetupTest()
	s.setupIsolation()
}

func (s *RedirTestSuite) TestTCP() {
	s.isolated.Client.SetupTCPRedirect(s.T(), redirPort, redir6Port, redirFwmark)
	s.isolated.Client.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-redir", fmt.Sprintf(":%v", redirPort), "-redir6", fmt.Sprintf(":%v", redir6Port), "-fwmark", fmt.Sprintf("0x%x", redirFwmark))
	})
	s.isolated.Server.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-tcp")
	})
	var peerPort uint16
	s.isolated.Peer.Do(s.T(), func() {
		peerPort = tcpEcho(s.T(), s.ctx)
	})
	s.Run("IPv4Peer", func() { s.testTCP(joinIPPort(s.peerIPv4, peerPort)) })
	s.Run("IPv6Peer", func() { s.testTCP(joinIPPort(s.peerIPv6, peerPort)) })
}

func (s *RedirTestSuite) testTCP(address string) {
	s.eventually(func(c *assert.CollectT) {
		conn, err := s.dial(&net.Dialer{}, "tcp", address)
		if !assert.NoError(c, err) {
			return
		}
		defer func() { assert.NoError(c, conn.Close()) }()

		r := bufio.NewReader(conn)
		for i := 0; i < 5; i++ {
			reqStr := fmt.Sprintf("%v\n", rand.Uint64())
			if _, err := conn.Write([]byte(reqStr)); !assert.NoError(c, err) {
				return
			}

			resStr, err := r.ReadString('\n')
			if !assert.NoError(c, err) {
				return
			}
			assert.Equal(c, fmt.Sprintf("ECHO %s", reqStr), resStr)
		}
	})
}
