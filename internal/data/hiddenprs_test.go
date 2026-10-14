package data

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHiddenPRStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hidden-prs.json")
	s := newHiddenPRStore(path)
	t.Cleanup(s.saving.Wait)
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	s.Hide("a", t0)
	s.Hide("b", t0)
	require.True(t, s.IsHidden("a", t0))

	require.False(t, s.IsHidden("a", t0.Add(time.Hour)), "comes back when it changes")
	require.False(t, s.IsHidden("a", t0), "entry is gone once it came back")

	s.Unhide("b")
	require.False(t, s.IsHidden("b", t0))

	s.Hide("c", t0)
	s.save()
	require.True(t, newHiddenPRStore(path).IsHidden("c", t0))
}
