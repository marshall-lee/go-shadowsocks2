//go:build !linux && !freebsd
package sockopt

import (
	"fmt"
	"runtime"
)

func SetRecvOrigDst(fd uintptr, ipv4 bool) error {
	return fmt.Errorf("original destination message not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
