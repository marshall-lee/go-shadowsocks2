//go:build darwin || freebsd
package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetDontFrag(fd uintptr, ipv4 bool) error {
	if ipv4 {
		return unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_DONTFRAG, 1)
	} else {
		return unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_DONTFRAG, 1)
	}
}
