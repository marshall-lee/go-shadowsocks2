//go:build linux
package sockopt

import (
	"golang.org/x/sys/unix"
)

func SetSocketMark(fd uintptr, mark uint) error {
	return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, int(mark))
}
