package udp

import (
	"context"
	"net"
	"net/netip"
	"syscall"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/sockopt"
)

type Conn interface {
	Empty() bool
	LocalAddr() netip.AddrPort
	SetReadDeadline(time.Time) error
	ReadFrom(buf []byte) (n int, addr netip.AddrPort, err error)
	ReadMsg(buf []byte) (n int, addr netip.AddrPort, msg Msg, err error)
	WriteTo(buf []byte, addr netip.AddrPort) (int, error)
	Close() error
}

type conn struct {
	udp *net.UDPConn
}

type ListenConfig struct {
	Transparent   bool
	RecvOrigDst   bool
	DoNotFragment bool
	ReusePort     bool
}

func (cfg ListenConfig) Listen(ctx context.Context, addr string) (Conn, error) {
	lc := net.ListenConfig{Control: cfg.Control}
	c, err := lc.ListenPacket(ctx, "udp", addr)
	if err != nil {
		return conn{}, err
	}
	return conn{udp: c.(*net.UDPConn)}, nil
}

func (cfg ListenConfig) Control(network, address string, c syscall.RawConn) error {
	var ctrlErr error
	ipv4 := network == "udp4"
	err := c.Control(func(fd uintptr) {
		if cfg.Transparent {
			if ctrlErr = sockopt.SetIPTransparent(fd, ipv4); ctrlErr != nil {
				return
			}
		}
		if cfg.RecvOrigDst {
			if ctrlErr = sockopt.SetRecvOrigDst(fd, ipv4); ctrlErr != nil {
				return
			}
		}
		if cfg.DoNotFragment {
			if ctrlErr = sockopt.SetDontFrag(fd, ipv4); ctrlErr != nil {
				return
			}
		}
		if cfg.ReusePort {
			if ctrlErr = sockopt.SetReusePort(fd); ctrlErr != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}
	return ctrlErr
}

func (c conn) Empty() bool {
	return c.udp == nil
}

func (c conn) LocalAddr() netip.AddrPort {
	return c.udp.LocalAddr().(*net.UDPAddr).AddrPort()
}

func (c conn) SetReadDeadline(t time.Time) error {
	return c.udp.SetReadDeadline(t)
}

func (c conn) ReadFrom(buf []byte) (n int, addr netip.AddrPort, err error) {
	return c.udp.ReadFromUDPAddrPort(buf)
}

func (c conn) ReadMsg(buf []byte) (n int, addr netip.AddrPort, msg Msg, err error) {
	var oob [64]byte
	n, oobn, _, addr, err := c.udp.ReadMsgUDPAddrPort(buf, oob[:])
	if err != nil {
		return n, addr, Msg{}, err
	}
	msg, err = parseMsg(oob[:oobn])
	return n, addr, msg, err
}

func (c conn) WriteTo(buf []byte, addr netip.AddrPort) (int, error) {
	return c.udp.WriteToUDPAddrPort(buf, addr)
}

func (c conn) Close() error {
	return c.udp.Close()
}
