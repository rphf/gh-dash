package tabs

import (
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/scopepicker"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

// viewRepoScope renders the repo in scope, or nothing when the scope is off.
func (m Model) viewRepoScope() string {
	if !m.ctx.RepoScoped() {
		return ""
	}
	repo := m.ctx.ScopeRepo()
	if repo == "" {
		repo = scopepicker.AllRepos
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(m.ctx.Styles.Tabs.TabSeparator.GetForeground()).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(m.ctx.Theme.PrimaryText).Bold(true).
				Render(" "+repo),
			lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).
				Render(" "+keys.Keys.RepoScope.Help().Key),
		))
}
