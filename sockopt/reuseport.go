//go:build !solaris && !windows
package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetReusePort(fd uintptr) error {
	return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}
