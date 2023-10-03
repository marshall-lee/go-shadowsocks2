package isolated

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

type link struct {
	netlink.Link
	ns namespace
}

func addVethPair(t *testing.T, ns namespace, name string, peerNS namespace, peerName string) (link, link) {
	t.Helper()
	var (
		veth, peer netlink.Link
		err        error
	)

	veth = &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:      name,
			Namespace: netlink.NsFd(ns.handle),
		},
		PeerName:      peerName,
		PeerNamespace: netlink.NsFd(peerNS.handle),
	}
	err = netlink.LinkAdd(veth)
	require.NoErrorf(t, err, "Failed to add %s link", name)

	ns.do(t, func() {
		veth, err = netlink.LinkByName(name)
		require.NoErrorf(t, err, "Failed to get %s link handle", name)
	})
	peerNS.do(t, func() {
		peer, err = netlink.LinkByName(peerName)
		require.NoErrorf(t, err, "Failed to get %s peer handle", peerName)
	})
	vethLink := link{veth, ns}
	peerLink := link{peer, peerNS}
	vethLink.setUp(t)
	peerLink.setUp(t)
	return vethLink, peerLink
}

func addBridge(t *testing.T, ns namespace, name string, links ...netlink.Link) link {
	t.Helper()
	var (
		bridge netlink.Link
		err    error
	)
	bridge = &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:      name,
			Namespace: netlink.NsFd(ns.handle),
		},
	}

	err = netlink.LinkAdd(bridge)
	require.NoErrorf(t, err, "Failed to add %s link", name)

	ns.do(t, func() {
		bridge, err = netlink.LinkByName(name)
		require.NoErrorf(t, err, "Failed to get %s link handle", name)
		for _, link := range links {
			require.NoErrorf(t,
				netlink.LinkSetMaster(link, bridge),
				"Failed to set %s master to %s", link.Attrs().Name, bridge.Attrs().Name,
			)
		}
	})
	link := link{bridge, ns}
	link.setUp(t)
	return link
}

func (link link) addAddr(t *testing.T, ip net.IP, mask net.IPMask) {
	t.Helper()
	if mask == nil {
		mask = ip.DefaultMask()
	}
	link.ns.do(t, func() {
		addr := netlink.Addr{IPNet: &net.IPNet{IP: ip, Mask: mask}}
		err := netlink.AddrAdd(link, &addr)
		require.NoErrorf(t, err, "Failed to add addr %s on %s", ip, link.Attrs().Name)
	})
}

func (link link) setUp(t *testing.T) {
	t.Helper()
	link.ns.do(t, func() {
		err := netlink.LinkSetUp(link)
		require.NoErrorf(t, err, "Failed to set %s link up", link.Attrs().Name)
	})
	t.Cleanup(func() {
		link.ns.do(t, func() {
			err := netlink.LinkSetDown(link)
			assert.NoErrorf(t, err, "Failed to set %s link down", link.Attrs().Name)
		})
	})
}
