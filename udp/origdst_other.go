//go:build !linux
package udp

import (
	"fmt"
	"net/netip"
	"runtime"
)

func (msg Msg) GetOrigDstAddr() (netip.AddrPort, error) {
	return netip.AddrPort{}, fmt.Errorf("original destination message not supported on %s-%s", runtime.GOOS, runtime.GOARCH)
}
