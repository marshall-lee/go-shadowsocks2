package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetIPTransparent(fd uintptr, ipv4 bool) error {
	return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_BINDANY, 1)
}
