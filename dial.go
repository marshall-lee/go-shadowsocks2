package main

import (
	"syscall"
	"net"

	"github.com/shadowsocks/go-shadowsocks2/sockopt"
)

func dial(network, address string) (net.Conn, error) {
	d := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			if err := c.Control(setDialOpts); err != nil {
				return err
			}
			return nil
		},
	}
	return d.Dial(network, address)
}

func setDialOpts(fd uintptr) {
	if config.Fwmark != 0 {
		if err := sockopt.SetSocketMark(fd, config.Fwmark); err != nil {
			logf("failed to set up dialing socket: %s", err)
		}
	}
}
