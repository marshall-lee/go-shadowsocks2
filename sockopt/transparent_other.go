//go:build !linux && !freebsd && !openbsd
package sockopt

import (
	"fmt"
	"runtime"
)

func SetIPTransparent(fd uintptr, ipv4 bool) error {
	return fmt.Errorf("transparent proxying not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
