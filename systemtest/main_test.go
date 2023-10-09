package systemtest

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	flag.StringVar(&shadowsocksPath, "shadowsocks-path", "", "Path to shadowsocks2 binary")
	flag.Parse()

	if shadowsocksPath == "" {
		fmt.Fprintln(os.Stderr, "You must provide -shadowsocks-path")
		return -1
	}

	runtime.GOMAXPROCS(2)
	return m.Run()
}
