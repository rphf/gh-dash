package gitspice

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasSpiceData(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	git("init", "-q")
	git("-c", "user.name=t", "-c", "user.email=t@example.com",
		"commit", "-q", "--allow-empty", "-m", "init")
	require.False(t, HasSpiceData(dir))

	git("update-ref", "refs/spice/data", "HEAD")
	require.True(t, HasSpiceData(dir))
}
