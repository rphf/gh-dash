package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/scopepicker"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

// updateRepoScope handles the repo scope picker, its keys and its messages.
// It tells if it handled msg.
func (m *Model) updateRepoScope(msg tea.Msg, currSection section.Section) (tea.Cmd, bool) {
	if !m.ctx.RepoScoped() {
		return nil, false
	}
	if m.scopePicker == nil {
		p := scopepicker.New(m.ctx)
		m.scopePicker = &p
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case scopepicker.TopReposMsg:
		*m.scopePicker, cmd = m.scopePicker.Update(msg)
		return cmd, true

	case scopepicker.SelectedMsg:
		return m.setRepoScope(msg.Repo), true

	case tea.KeyPressMsg:
		if m.scopePicker.IsOpen() {
			*m.scopePicker, cmd = m.scopePicker.Update(msg)
			return cmd, true
		}
		if m.isTyping(msg, currSection) {
			return nil, false
		}
		switch {
		case key.Matches(msg, m.keys.RepoScope):
			return m.scopePicker.Open(), true
		case m.isScopeToggleKey(msg):
			repo := m.ctx.ToggledScopeRepo()
			if repo == "" && m.ctx.ScopeRepo() == "" {
				return m.scopePicker.Open(), true
			}
			return m.setRepoScope(repo), true
		}
	}
	return nil, false
}

// isTyping tells if a key goes to an input, a prompt or a pending action.
func (m *Model) isTyping(msg tea.KeyPressMsg, currSection section.Section) bool {
	return (currSection != nil &&
		(currSection.IsSearchFocused() || currSection.IsPromptConfirmationFocused())) ||
		m.prView.IsTextInputBoxFocused() ||
		m.issueSidebar.IsTextInputBoxFocused() ||
		m.footer.ShowConfirmQuit ||
		m.notificationView.HasPendingAction() ||
		m.mode == ModeSection ||
		m.isUserDefinedKeybinding(msg)
}

// isScopeToggleKey tells if msg is the smart filtering key of the view,
// which toggles between the last repo and all repos.
func (m *Model) isScopeToggleKey(msg tea.KeyPressMsg) bool {
	switch m.ctx.View {
	case config.PRsView:
		return key.Matches(msg, keys.PRKeys.ToggleSmartFiltering)
	case config.IssuesView:
		return key.Matches(msg, keys.IssueKeys.ToggleSmartFiltering)
	case config.NotificationsView:
		return key.Matches(msg, keys.NotificationKeys.ToggleSmartFiltering)
	}
	return false
}

// setRepoScope puts repo in scope, or all repos when it is empty, and
// fetches the sections again.
func (m *Model) setRepoScope(repo string) tea.Cmd {
	if strings.EqualFold(repo, m.ctx.ScopeRepo()) {
		return nil
	}
	m.ctx.SetScopeRepo(repo)

	for _, sections := range [][]section.Section{m.prs, m.issues, m.notifications} {
		for _, s := range sections {
			if s == nil {
				continue
			}
			s.ResetRows()
			if st, ok := s.(interface{ MarkStale() }); ok {
				st.MarkStale()
			}
		}
	}

	m.tabs.UpdateProgramContext(m.ctx)
	fetch := m.fetchAllViewSections()
	m.updateTabs()
	return tea.Batch(fetch, m.onViewedRowChanged(), saveRecentRepo(repo))
}

func saveRecentRepo(repo string) tea.Cmd {
	return func() tea.Msg {
		if err := data.AddRecentRepo(repo); err != nil {
			log.Warn("failed to save recent repo", "err", err)
		}
		return nil
	}
}

// repoScopeLayers returns the repo scope picker as a layer under the header.
func (m *Model) repoScopeLayers() []*lipgloss.Layer {
	if m.scopePicker == nil || !m.scopePicker.IsOpen() {
		return nil
	}
	return []*lipgloss.Layer{
		lipgloss.NewLayer(m.scopePicker.View()).X(1).Y(common.HeaderHeight).Z(1),
	}
}
