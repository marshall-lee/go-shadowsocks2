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
	tunPort  = 1090
	tun6Port = 1091
)

type TunTestSuite struct {
	SystemTestSuite
}

func TestTun(t *testing.T) {
	suite.Run(t, new(TunTestSuite))
}

func (s *TunTestSuite) TestTCP() {
	peerPort := tcpEcho(s.T(), s.ctx)
	peerHostPort := joinIPPort(s.peerIPv4, peerPort)
	peer6HostPort := joinIPPort(s.peerIPv6, peerPort)
	s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-tcptun", fmt.Sprintf(":%v=%s,:%v=%s", tunPort, peerHostPort, tun6Port, peer6HostPort))
	s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-tcp")
	s.Run("IPv4Peer", func() { s.testTCP(joinIPPort(s.clientIPv4, tunPort)) })
	s.Run("IPv6Peer", func() { s.testTCP(joinIPPort(s.clientIPv6, tun6Port)) })
}

func (s *TunTestSuite) TestTCPIsolated() {
	s.setupIsolation()

	var (
		peerHostPort  string
		peer6HostPort string
	)
	s.isolated.Peer.Do(s.T(), func() {
		peerPort := tcpEcho(s.T(), s.ctx)
		peerHostPort = joinIPPort(s.peerIPv4, peerPort)
		peer6HostPort = joinIPPort(s.peerIPv6, peerPort)
	})
	s.isolated.Client.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-tcptun", fmt.Sprintf(":%v=%s,:%v=%s", tunPort, peerHostPort, tun6Port, peer6HostPort))
	})
	s.isolated.Server.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-tcp")
	})
	s.Run("IPv4Peer", func() { s.testTCP(joinIPPort(s.clientIPv4, tunPort)) })
	s.Run("IPv6Peer", func() { s.testTCP(joinIPPort(s.clientIPv6, tun6Port)) })
}

func (s *TunTestSuite) testTCP(address string) {
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

func (s *TunTestSuite) TestUDP() {
	peerPort := udpEcho(s.T(), s.ctx)
	peerHostPort := joinIPPort(s.peerIPv4, peerPort)
	peer6HostPort := joinIPPort(s.peerIPv6, peerPort)
	s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-udptun", fmt.Sprintf(":%v=%s,:%v=%s", tunPort, peerHostPort, tun6Port, peer6HostPort))
	s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-udp")
	s.Run("IPv4Peer", func() { s.testUDP(joinIPPort(s.clientIPv4, tunPort)) })
	s.Run("IPv6Peer", func() { s.testUDP(joinIPPort(s.clientIPv6, tun6Port)) })
}

func (s *TunTestSuite) TestUDPIsolated() {
	s.setupIsolation()

	var (
		peerHostPort  string
		peer6HostPort string
	)
	s.isolated.Peer.Do(s.T(), func() {
		peerPort := udpEcho(s.T(), s.ctx)
		peerHostPort = joinIPPort(s.peerIPv4, peerPort)
		peer6HostPort = joinIPPort(s.peerIPv6, peerPort)
	})
	s.isolated.Client.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-c", s.serverHostPort(), "-udptun", fmt.Sprintf(":%v=%s,:%v=%s", tunPort, peerHostPort, tun6Port, peer6HostPort))
	})
	s.isolated.Server.Do(s.T(), func() {
		s.shadowsocks("-verbose", "-password", "secretpwd", "-s", fmt.Sprintf(":%v", shadowsocksPort), "-udp")
	})
	s.Run("IPv4Peer", func() { s.testUDP(joinIPPort(s.clientIPv4, tunPort)) })
	s.Run("IPv6Peer", func() { s.testUDP(joinIPPort(s.clientIPv6, tun6Port)) })
}

func (s *TunTestSuite) testUDP(address string) {
	s.eventually(func(c *assert.CollectT) {
		conn, err := s.dial(&net.Dialer{}, "udp", address)
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
