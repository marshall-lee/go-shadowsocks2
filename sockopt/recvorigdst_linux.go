package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetRecvOrigDst(fd uintptr, ipv4 bool) error {
	if ipv4 {
		return unix.SetsockoptInt(int(fd), unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
	} else {
		return unix.SetsockoptInt(int(fd), unix.SOL_IPV6, unix.IPV6_RECVORIGDSTADDR, 1)
	}
}
