package udp

import (
	"net/netip"

	"golang.org/x/sys/unix"
)

func (msg Msg) GetOrigDstAddr() (netip.AddrPort, error) {
	for _, cmsg := range msg.cmsgs {
		sockaddr, err := unix.ParseOrigDstAddr(&cmsg)
		if err != nil {
			continue
		}
		switch sockaddr := sockaddr.(type) {
		case *unix.SockaddrInet4:
			return netip.AddrPortFrom(netip.AddrFrom4(sockaddr.Addr), uint16(sockaddr.Port)), nil
		case *unix.SockaddrInet6:
			return netip.AddrPortFrom(netip.AddrFrom16(sockaddr.Addr), uint16(sockaddr.Port)), nil
		}
	}
	return netip.AddrPort{}, unix.EINVAL
}
