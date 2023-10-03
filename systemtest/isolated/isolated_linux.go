package isolated

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type Node struct {
	ns         namespace
	ipv4, ipv6 net.IP
}

type Isolated struct {
	Client, Server, Peer Node
}

func New(t *testing.T) *Isolated {
	isolated := Isolated{
		Client: Node{
			ns:   newNS(t),
			ipv4: net.IPv4(192, 168, 1, 1),
			ipv6: net.IP([]byte{0xfd, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}),
		},
		Server: Node{
			ns:   newNS(t),
			ipv4: net.IPv4(192, 168, 1, 2),
			ipv6: net.IP([]byte{0xfd, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}),
		},
		Peer: Node{
			ns:   newNS(t),
			ipv4: net.IPv4(192, 168, 1, 3),
			ipv6: net.IP([]byte{0xfd, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03}),
		},
	}

	// Set up virtual network interfaces.
	veth0, veth1 := addVethPair(t, isolated.Client.ns, "veth0", isolated.Server.ns, "veth1")
	veth2, veth3 := addVethPair(t, isolated.Server.ns, "veth2", isolated.Peer.ns, "veth3")
	br0 := addBridge(t, isolated.Server.ns, "br0", veth1, veth2)

	// Assign ipv4 addresses.
	veth0.addAddr(t, isolated.Client.ipv4, nil)
	veth0.addAddr(t, isolated.Client.ipv6, net.CIDRMask(64, 128))

	br0.addAddr(t, isolated.Server.ipv4, nil)
	br0.addAddr(t, isolated.Server.ipv6, net.CIDRMask(64, 128))

	veth3.addAddr(t, isolated.Peer.ipv4, nil)
	veth3.addAddr(t, isolated.Peer.ipv6, net.CIDRMask(64, 128))

	// Use client namespace by default.
	isolated.Client.ns.set(t)

	return &isolated
}

func (n Node) IPv4() net.IP {
	return n.ipv4
}

func (n Node) IPv6() net.IP {
	return n.ipv6
}

func (n Node) Do(t require.TestingT, f func()) {
	n.ns.do(t, f)
}

