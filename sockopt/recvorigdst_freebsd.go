package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetRecvOrigDst(fd uintptr, ipv4 bool) error {
	if ipv4 {
		return unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_ORIGDSTADDR, 1)
	} else {
		return unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_ORIGDSTADDR, 1)
	}
}
