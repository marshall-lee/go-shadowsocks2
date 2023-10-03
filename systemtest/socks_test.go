package systemtest

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/proxy"
)

const socksPort = 1080

type SocksTestSuite struct {
	SystemTestSuite
}

func TestSocks(t *testing.T) {
	suite.Run(t, new(SocksTestSuite))
}

func (s *SocksTestSuite) TestTCP() {
	s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-socks", fmt.Sprintf(":%v", socksPort))
	s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-tcp")
	peerPort := tcpEcho(s.T(), s.ctx)

	s.Run("IPv4Peer", func() { s.testTCP(joinIPPort(s.peerIPv4, peerPort)) })
	s.Run("IPv6Peer", func() { s.testTCP(joinIPPort(s.peerIPv6, peerPort)) })
}

func (s *SocksTestSuite) TestTCPIsolated() {
	s.setupIsolation()

	s.isolated.Client.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-socks", fmt.Sprintf(":%v", socksPort))
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

func (s *SocksTestSuite) testTCP(address string) {
	s.eventually(func(c *assert.CollectT) {
		conn, err := s.dialSocks(c, "tcp", s.socksHostPort(), "tcp", address)
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

func (s *SocksTestSuite) TestUDP() {
	s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-socks", fmt.Sprintf(":%v", socksPort), "-u")
	s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-udp")
	peerPort := udpEcho(s.T(), s.ctx)

	s.Run("IPv4Peer", func() { s.testUDP(joinIPPort(s.peerIPv4, peerPort)) })
	s.Run("IPv6Peer", func() { s.testUDP(joinIPPort(s.peerIPv6, peerPort)) })
}

func (s *SocksTestSuite) TestUDPIsolated() {
	s.setupIsolation()

	s.isolated.Client.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-socks", fmt.Sprintf(":%v", socksPort), "-u")
	})
	s.isolated.Server.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-udp")
	})
	var peerPort uint16
	s.isolated.Peer.Do(s.T(), func() {
		peerPort = udpEcho(s.T(), s.ctx)
	})

	s.Run("IPv4Peer", func() { s.testUDP(joinIPPort(s.peerIPv4, peerPort)) })
	s.Run("IPv6Peer", func() { s.testUDP(joinIPPort(s.peerIPv6, peerPort)) })
}

func (s *SocksTestSuite) testUDP(address string) {
	s.eventually(func(c *assert.CollectT) {
		conn, err := s.dialSocks(c, "tcp", s.socksHostPort(), "udp", address)
		if !assert.NoError(c, err) {
			return
		}
		defer func() { assert.NoError(c, conn.Close()) }()

		var buf [32]byte
		for i := 0; i < 10; i++ {
			reqStr := fmt.Sprintf("%v", rand.Uint64())
			_, err := conn.Write([]byte(reqStr))
			if !assert.NoError(c, err) {
				return
			}

			n, err := conn.Read(buf[:])
			if !assert.NoError(c, err) {
				return
			}
			resStr := string(buf[:n])
			assert.Equal(c, fmt.Sprintf("ECHO %s", reqStr), resStr)
		}
	})
}

func (s *SocksTestSuite) socksHostPort() string {
	return joinIPPort(s.clientIPv4, socksPort)
}

func (s *SocksTestSuite) dialSocks(t assert.TestingT, socksNetwork, socksAddress, network, address string) (net.Conn, error) {
	dialer, err := proxy.SOCKS5(socksNetwork, socksAddress, nil, nil)
	if !assert.NoError(t, err) {
		return nil, err
	}

	return s.dial(dialer.(proxy.ContextDialer), network, address)
}
