//go:build linux || freebsd

package child

import (
	"io"
	"testing"
	"os/exec"
	"syscall"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func Start(t *testing.T, output io.Writer, name string, arg ...string) {
	t.Helper()
	cmd := exec.Command(name, arg...)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}
	err := cmd.Start()
	require.NoError(t, err, "Failed to start child process")
	t.Cleanup(func() {
		err := cmd.Process.Kill()
		assert.NoError(t, err, "Failed to kill child process")
	})
}
