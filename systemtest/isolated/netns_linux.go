package isolated

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type namespace struct {
	handle netns.NsHandle
}

func newNS(t *testing.T) namespace {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	orighandle, err := netns.Get()
	require.NoError(t, err, "Failed to get network namespace")
	defer func() {
		require.NoError(t, netns.Set(orighandle), "Failed to restore the original network namespace")
		require.NoError(t, orighandle.Close(), "Failed to close the original network namespace")
	}()

	handle, err := netns.New()
	require.NoError(t, err, "Failed to create network namespace")
	t.Cleanup(func() {
		assert.NoError(t, handle.Close(), "Failed to close namespace handle")
	})

	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err, "Failed to get loopback interface")

	err = netlink.LinkSetUp(lo)
	require.NoError(t, err, "Failed to set loopback interface up")

	ns := namespace{handle}
	ns.nftrace(t)
	return ns
}

func (ns namespace) set(t *testing.T) {
	t.Helper()
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)

	orighandle, err := netns.Get()
	require.NoError(t, err, "Failed to get network namespace")

	t.Cleanup(func() {
		assert.NoError(t, netns.Set(orighandle), "Failed to restore the original network namespace")
		assert.NoError(t, orighandle.Close(), "Failed to close the original network namespace")
	})

	require.NoError(t, netns.Set(ns.handle), "Failed to set network namespace")
}

func (ns namespace) do(t require.TestingT, f func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	orighandle, err := netns.Get()
	require.NoError(t, err, "Failed to get network namespace")
	defer func() {
		require.NoError(t, netns.Set(orighandle), "Failed to restore the original network namespace")
		require.NoError(t, orighandle.Close(), "Failed to close the original network namespace")
	}()

	require.NoError(t, netns.Set(ns.handle), "Failed to set network namespace")

	f()
}
