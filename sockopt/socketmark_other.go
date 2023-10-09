//go:build !linux && !freebsd

package sockopt

import (
	"fmt"
	"runtime"
)

func SetSocketMark(fd uintptr, mark uint) error {
	return fmt.Errorf("socket mark not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
