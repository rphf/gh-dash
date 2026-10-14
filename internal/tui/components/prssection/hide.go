package prssection

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

// withoutHidden records the hidden PRs of prs and returns prs without them,
// or prs when the hidden PRs are shown. It reuses the array of prs.
func (m *Model) withoutHidden(prs []prrow.Data) []prrow.Data {
	store := data.GetHiddenPRStore()
	m.rows.hidden = map[string]bool{}
	res := prs[:0]
	for _, pr := range prs {
		if store.IsHidden(pr.Primary.Url, pr.Primary.UpdatedAt) {
			m.rows.hidden[pr.Primary.Url] = true
			if !m.rows.showHidden {
				continue
			}
		}
		res = append(res, pr)
	}
	return res
}

// withHiddenCount adds the number of hidden PRs to the pager content, with
// the key that shows or hides them.
func (m Model) withHiddenCount(content string) string {
	n := len(m.rows.hidden)
	if n == 0 || content == "" {
		return content
	}
	action := "show"
	if m.rows.showHidden {
		action = "hide"
	}
	return fmt.Sprintf("%s • %d hidden (%s to %s)",
		content, n, keys.PRKeys.ShowHidden.Help().Key, action)
}

// toggleHidden hides the current PR, or unhides it when it is hidden. A
// hidden PR comes back when it changes. The PRs stacked on a root are hidden
// and unhidden with it.
func (m *Model) toggleHidden() {
	idx := m.Table.GetCurrItem()
	if idx < 0 || idx >= len(m.Prs) {
		return
	}
	pr := m.Prs[idx].Primary
	store := data.GetHiddenPRStore()
	unhide := store.IsHidden(pr.Url, pr.UpdatedAt)

	for _, t := range m.rows.all {
		u := t.Primary.Url
		if u != pr.Url && m.rows.roots[u] != pr.Url {
			continue
		}
		if unhide {
			store.Unhide(u)
		} else {
			store.Hide(u, t.Primary.UpdatedAt)
		}
	}

	m.arrangeStacks(false)
	for m.Table.GetCurrItem() >= len(m.Prs) && m.Table.GetCurrItem() > 0 {
		m.Table.PrevItem()
	}
	m.Table.SetRows(m.BuildRows())
}

// toggleShowHidden shows or hides the hidden PRs.
func (m *Model) toggleShowHidden() {
	m.rows.showHidden = !m.rows.showHidden
	m.arrangeStacks(false)
	for m.Table.GetCurrItem() >= len(m.Prs) && m.Table.GetCurrItem() > 0 {
		m.Table.PrevItem()
	}
	m.Table.SetRows(m.BuildRows())
}

// dimLines renders lines in the faint text color, on the selected background
// when selected.
func dimLines(ctx *context.ProgramContext, lines []string, selected bool) {
	style := lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	if selected {
		style = style.Background(ctx.Theme.SelectedBackground)
	}
	for i, line := range lines {
		lines[i] = style.Render(ansi.Strip(line))
	}
}
