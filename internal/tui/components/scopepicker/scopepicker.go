// Package scopepicker is the popup that picks the repo in scope.
package scopepicker

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

const (
	// AllRepos is the label of the scope with all repos.
	AllRepos   = "All repos"
	maxVisible = 12
	topRepos   = 30
)

// TopReposMsg carries the repos the viewer contributes to.
type TopReposMsg struct {
	Repos []string
}

// SelectedMsg is sent when a scope is picked. An empty repo means all repos.
type SelectedMsg struct {
	Repo string
}

type Model struct {
	ctx      *context.ProgramContext
	open     bool
	input    textinput.Model
	matches  []string
	cursor   int
	topRepos []string
	fetched  bool
}

func New(ctx *context.ProgramContext) Model {
	ti := textinput.New()
	ti.Prompt = "  "
	ti.Placeholder = "owner/repo"
	return Model{ctx: ctx, input: ti}
}

func (m Model) IsOpen() bool {
	return m.open
}

// Open shows the picker with the cursor on the repo in scope.
func (m *Model) Open() tea.Cmd {
	m.open = true
	m.input.SetValue("")
	m.styleInput()
	m.filter()
	m.cursor = 0
	for i, repo := range m.matches {
		if strings.EqualFold(repo, m.ctx.ScopeRepo()) {
			m.cursor = i
		}
	}

	cmds := []tea.Cmd{m.input.Focus()}
	if !m.fetched {
		m.fetched = true
		cmds = append(cmds, fetchTopRepos)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Close() {
	m.open = false
	m.input.Blur()
}

func fetchTopRepos() tea.Msg {
	repos, err := data.FetchTopRepos(topRepos)
	if err != nil {
		log.Warn("failed to fetch top repos", "err", err)
	}
	return TopReposMsg{Repos: repos}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TopReposMsg:
		m.topRepos = msg.Repos
		if m.open {
			m.filter()
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.open {
			return m, nil
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			m.Close()
			return m, nil
		case "enter":
			repo, ok := m.selection()
			if !ok {
				return m, nil
			}
			m.Close()
			return m, func() tea.Msg { return SelectedMsg{Repo: repo} }
		case "up", "ctrl+p", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n", "ctrl+j":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
			}
			return m, nil
		}
		var cmd tea.Cmd
		before := m.input.Value()
		m.input, cmd = m.input.Update(msg)
		if m.input.Value() != before {
			m.filter()
			m.cursor = 0
		}
		return m, cmd
	}
	return m, nil
}

// selection returns the repo under the cursor, or the typed repo when
// nothing matches it.
func (m Model) selection() (string, bool) {
	if m.cursor < len(m.matches) {
		return m.matches[m.cursor], true
	}
	typed := strings.TrimSpace(m.input.Value())
	if owner, name, ok := strings.Cut(typed, "/"); ok && owner != "" && name != "" &&
		!strings.ContainsAny(typed, " \t") {
		return typed, true
	}
	return "", false
}

// candidates returns the repos to pick from, in order: all repos, the repo
// gh-dash runs in, the recent repos, the repos the viewer contributes to and
// the repos seen in rows.
func (m Model) candidates() []string {
	seen := map[string]bool{}
	repos := []string{""}
	add := func(list ...string) {
		for _, r := range list {
			k := strings.ToLower(r)
			if r != "" && !seen[k] {
				seen[k] = true
				repos = append(repos, r)
			}
		}
	}
	add(m.ctx.ScopeRepo(), m.ctx.HomeRepo())
	add(data.RecentRepos()...)
	add(m.topRepos...)
	add(m.ctx.KnownRepos()...)
	return repos
}

func (m *Model) filter() {
	terms := strings.Fields(strings.ToLower(m.input.Value()))
	m.matches = nil
	for _, repo := range m.candidates() {
		label := strings.ToLower(label(repo))
		ok := true
		for _, t := range terms {
			if !strings.Contains(label, t) {
				ok = false
				break
			}
		}
		if ok {
			m.matches = append(m.matches, repo)
		}
	}
}

func label(repo string) string {
	if repo == "" {
		return AllRepos
	}
	return repo
}

func (m *Model) styleInput() {
	theme := m.ctx.Theme
	m.input.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Placeholder: lipgloss.NewStyle().Foreground(theme.FaintText),
			Prompt:      lipgloss.NewStyle().Foreground(theme.SecondaryText),
			Text:        lipgloss.NewStyle().Foreground(theme.PrimaryText),
		},
		Cursor: textinput.CursorStyle{
			Color: theme.FaintText,
			Shape: tea.CursorBar,
			Blink: false,
		},
	})
}

func (m Model) width() int {
	w := 40
	for _, repo := range m.matches {
		w = max(w, lipgloss.Width(label(repo))+6)
	}
	return min(w, max(m.ctx.ScreenWidth-4, 20))
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	theme := m.ctx.Theme
	width := m.width()
	m.input.SetWidth(width - 4)

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := min(start+maxVisible, len(m.matches))

	current := m.ctx.ScopeRepo()
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryText).Render(" Switch repo"),
		m.input.View(),
	}
	for i := start; i < end; i++ {
		repo := m.matches[i]
		mark := "  "
		if strings.EqualFold(repo, current) {
			mark = "● "
		}
		style := lipgloss.NewStyle().Foreground(theme.SecondaryText)
		pointer := "  "
		if i == m.cursor {
			style = lipgloss.NewStyle().Foreground(theme.PrimaryText).Bold(true).
				Background(theme.SelectedBackground)
			pointer = "› "
		}
		lines = append(lines, style.Width(width-2).Render(pointer+mark+label(repo)))
	}
	if len(m.matches) == 0 {
		hint := "No repo matches"
		if _, ok := m.selection(); ok {
			hint = "enter: switch to " + strings.TrimSpace(m.input.Value())
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FaintText).Italic(true).
			Render("  "+hint))
	} else if hidden := len(m.matches) - (end - start); hidden > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.FaintText).
			Render(fmt.Sprintf("  %d more", hidden)))
	}

	return m.ctx.Styles.Select.PopupStyle.
		Width(width).
		Render(strings.Join(lines, "\n"))
}
