package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetIPTransparent(fd uintptr, ipv4 bool) error {
	if ipv4 {
		return unix.SetsockoptInt(int(fd), unix.SOL_IP, unix.IP_TRANSPARENT, 1)
	} else {
		return unix.SetsockoptInt(int(fd), unix.SOL_IPV6, unix.IPV6_TRANSPARENT, 1)
	}
}
