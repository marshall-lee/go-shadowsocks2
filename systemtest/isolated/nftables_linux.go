package isolated

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/shadowsocks/go-shadowsocks2/systemtest/child"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

var nftraceEnabled bool

func init() {
	flag.BoolVar(&nftraceEnabled, "nftrace", false, "Starts nft monitor trace in backgrorund")
}

func (ns namespace) nftConn(t *testing.T) (conn *nftables.Conn) {
	ns.do(t, func() {
		var err error
		conn, err = nftables.New(nftables.WithNetNSFd(int(ns.handle)))
		require.NoError(t, err, "Failed to establish netfilter connection")
	})
	return
}

// nftrace starts "nft monitor trace" background process to debug netfilter rules.
func (ns namespace) nftrace(t *testing.T) {
	t.Helper()
	if nftraceEnabled {
		ns.do(t, func() { child.Start(t, os.Stderr, "nft", "monitor", "trace") })
	}
}

/*
BanSourceIP drops any inbound packet from a specified IP address.

Example:

	table ip ban-source-192.168.1.1 {
		chain input {
			type filter hook input priority filter; policy accept;
			meta nftrace set 1
			ip saddr 192.168.1.1 drop
		}
	}

	table ip6 ban-source-fd11:1111:1111:1111::1 {
		chain input {
			type filter hook input priority filter; policy accept;
			meta nftrace set 1
			ip6 saddr fd11:1111:1111:1111::1 drop
		}
	}
*/
func (n Node) BanSourceIP(t *testing.T, ip net.IP) {
	nft := n.ns.nftConn(t)
	defer func() {
		require.NoError(t, nft.CloseLasting())
	}()
	var (
		payload expr.Payload
		family  nftables.TableFamily
	)
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
		family = nftables.TableFamilyIPv4
		payload = expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       12,
			Len:          4,
		}
	} else {
		family = nftables.TableFamilyIPv6
		payload = expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       8,
			Len:          16,
		}
	}
	table := nft.AddTable(&nftables.Table{
		Family: family,
		Name:   fmt.Sprintf("ban-source-%s", ip),
	})
	polAccept := nftables.ChainPolicyAccept
	input := nft.AddChain(&nftables.Chain{
		Name:     "input",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &polAccept,
	})

	nft.AddRule(nftraceRule(table, input))
	nft.AddRule(&nftables.Rule{
		Table: table,
		Chain: input,
		Exprs: []expr.Any{
			&payload,
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     ip,
			},
			&expr.Verdict{
				Kind: expr.VerdictDrop,
			},
		},
	})
	require.NoError(t, nft.Flush())
}

/*
SetupTCPRedirect establishes a redirect of outbound TCP connection to a specified port.
Outbound connections of the proxy itself are bypassed based on a packet mark.

Example:

	table inet tcp-redirect-1082/1083 {
		chain output {
			type nat hook output priority mangle; policy accept;
			meta nftrace set 1
			meta l4proto tcp meta mark 0x00001234 return
			meta nfproto ipv4 meta l4proto tcp redirect to :1082
			meta nfproto ipv6 meta l4proto tcp redirect to :1083
		}
	}
*/
func (n Node) SetupTCPRedirect(t *testing.T, redirPort, redir6Port uint16, mark uint32) {
	nft := n.ns.nftConn(t)
	defer func() {
		require.NoError(t, nft.CloseLasting())
	}()
	table := nft.AddTable(&nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   fmt.Sprintf("tcp-redirect-%v/%v", redirPort, redir6Port),
	})
	polAccept := nftables.ChainPolicyAccept
	output := nft.AddChain(&nftables.Chain{
		Name:     "output",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityMangle,
		Policy:   &polAccept,
	})
	nft.AddRule(nftraceRule(table, output))
	markBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(markBytes, mark)
	nft.AddRule(&nftables.Rule{
		Table: table,
		Chain: output,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_TCP},
			},
			&expr.Meta{Key: expr.MetaKeyMARK, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     markBytes,
			},
			&expr.Verdict{
				Kind: expr.VerdictReturn,
			},
		},
	})
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, redirPort)
	nft.AddRule(&nftables.Rule{
		Table: table,
		Chain: output,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.NFPROTO_IPV4},
			},
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_TCP},
			},
			&expr.Immediate{
				Register: 1,
				Data:     portBytes,
			},
			&expr.Redir{
				RegisterProtoMin: 1,
			},
		},
	})
	binary.BigEndian.PutUint16(portBytes, redir6Port)
	nft.AddRule(&nftables.Rule{
		Table: table,
		Chain: output,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.NFPROTO_IPV6},
			},
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_TCP},
			},
			&expr.Immediate{
				Register: 1,
				Data:     portBytes,
			},
			&expr.Redir{
				RegisterProtoMin: 1,
			},
		},
	})
	require.NoError(t, nft.Flush())
}

func nftraceRule(table *nftables.Table, chain *nftables.Chain) *nftables.Rule {
	oneBytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(oneBytes, 1)
	return &nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Immediate{
				Register: 1,
				Data:     oneBytes,
			},
			&expr.Meta{Key: expr.MetaKeyNFTRACE, Register: 1, SourceRegister: true},
		},
	}
}
