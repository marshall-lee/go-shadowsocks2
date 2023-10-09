//go:build !linux && !freebsd && !openbsd && !netbsd && !darwin

package child

import (
	"io"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Start(t *testing.T, output io.Writer, name string, arg ...string) {
	t.Helper()
	cmd := exec.Command(name, arg...)
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Start()
	require.NoError(t, err, "Failed to start child process")
	t.Cleanup(func() {
		err := cmd.Process.Kill()
		assert.NoError(t, err, "Failed to kill child process")
	})
}
