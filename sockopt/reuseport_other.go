//go:build solaris || windows

package sockopt

import (
	"fmt"
	"runtime"
)

func SetReusePort(fd uintptr) error {
	return fmt.Errorf("port reusage not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
