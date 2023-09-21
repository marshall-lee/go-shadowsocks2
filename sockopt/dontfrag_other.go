//go:build !linux && !darwin && !freebsd

package sockopt

import (
	"fmt"
	"runtime"
)

func SetDontFrag(fd uintptr, ipv4 bool) error {
	return fmt.Errorf("setting don't fragment flag not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
