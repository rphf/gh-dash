package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStacksEnabled(t *testing.T) {
	parse := func(t *testing.T, content string) Config {
		t.Helper()
		cfgPath := filepath.Join(t.TempDir(), "config.yml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
		cfg, err := ParseConfig(Location{ConfigFlag: cfgPath, SkipGlobalConfig: true})
		require.NoError(t, err)
		return cfg
	}

	t.Run("true when not set", func(t *testing.T) {
		require.True(t, parse(t, "confirmQuit: false\n").Stacks.IsEnabled())
	})
	t.Run("false when disabled", func(t *testing.T) {
		require.False(t, parse(t, "stacks:\n  enabled: false\n").Stacks.IsEnabled())
	})
}
