//go:build !linux && !darwin && !windows
// +build !linux,!darwin,!windows

package main

import (
	"net"
	"runtime"
)

func redirLocal(addr, server string, shadow func(net.Conn) net.Conn) {
	logf("TCP redirect not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}

func redir6Local(addr, server string, shadow func(net.Conn) net.Conn) {
	logf("TCP6 redirect not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}

// tcpSetListenOpts sets listening socket options.
func tcpSetListenOpts(fd uintptr) error {
	if config.TCPFastOpen {
		return fmt.Errorf("tcp-fast-open is not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
	}
	return nil
}

// tcpSetDialOpts sets dialing socket options.
func tcpSetDialOpts(fd uintptr) error {
	if config.TCPFastOpen {
		return fmt.Errorf("tcp-fast-open is not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
	}
	return nil
}
