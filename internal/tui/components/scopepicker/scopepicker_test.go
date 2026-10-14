package scopepicker

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newPicker(t *testing.T) Model {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	ctx := &context.ProgramContext{
		Theme:       *theme.DefaultTheme,
		Config:      &config.Config{},
		ScreenWidth: 100,
	}
	ctx.Styles = context.InitStyles(ctx.Theme)
	ctx.InitRepoScope()
	ctx.AddKnownRepos("o/app", "o/lib")
	m := New(ctx)
	m.fetched = true
	m.Open()
	return m
}

func typeText(m Model, s string) Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func pick(t *testing.T, m Model) string {
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg, ok := cmd().(SelectedMsg)
	require.True(t, ok)
	return msg.Repo
}

func TestCandidates(t *testing.T) {
	m := newPicker(t)
	m, _ = m.Update(TopReposMsg{Repos: []string{"o/lib", "o/web"}})
	require.Equal(t, []string{"", "o/lib", "o/web", "o/app"}, m.matches)
}

func TestFilterAndPick(t *testing.T) {
	m := typeText(newPicker(t), "lib")
	require.Equal(t, []string{"o/lib"}, m.matches)
	require.Equal(t, "o/lib", pick(t, m))

	m = typeText(newPicker(t), "all")
	require.Equal(t, "", pick(t, m))
}

func TestPickTypedRepo(t *testing.T) {
	m := typeText(newPicker(t), "x/y")
	require.Empty(t, m.matches)
	require.Equal(t, "x/y", pick(t, m))

	m = typeText(newPicker(t), "nothing")
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Nil(t, cmd)
}

func TestEscCloses(t *testing.T) {
	m := newPicker(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.False(t, m.IsOpen())
	require.Empty(t, m.View())
}
