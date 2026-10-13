package prssection

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/gitspice"
	"github.com/dlvhdr/gh-dash/v4/internal/stack"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func stacksEnabled(ctx *context.ProgramContext) bool {
	return ctx != nil && ctx.Config != nil && ctx.Config.Stacks.IsEnabled()
}

// loadStackSources fills the stack sources of prs from GitHub native stacks and
// from git-spice in the repo gh-dash runs in. It runs in the fetch command.
func loadStackSources(ctx *context.ProgramContext, prs []prrow.Data) {
	if !stacksEnabled(ctx) || len(prs) == 0 {
		return
	}

	urls := make([]string, 0, len(prs))
	for _, pr := range prs {
		urls = append(urls, pr.Primary.Url)
	}
	natives, err := data.FetchPullRequestStacks(urls)
	if err != nil {
		log.Warn("failed to fetch native PR stacks", "err", err)
	}

	spiceParents := map[string]map[string]string{}
	for i := range prs {
		pr := prs[i].Primary
		if native, ok := natives[pr.Url]; ok {
			prs[i].Stack.Sources.Native = &native
		}

		repo := pr.Repository.NameWithOwner
		parents, ok := spiceParents[repo]
		if !ok {
			parents = map[string]string{}
			if dir := localRepoPath(ctx, repo); dir != "" {
				parents = gitspice.ParentURLs(dir)
			}
			spiceParents[repo] = parents
		}
		prs[i].Stack.Sources.SpiceParentURL = parents[pr.Url]
	}
}

// localRepoPath returns the repo gh-dash runs in when it is a clone of repo
// tracked by git-spice.
func localRepoPath(ctx *context.ProgramContext, repo string) string {
	if ctx.HasGHRepo() && ctx.RepoPath != "" &&
		strings.EqualFold(ctx.GHRepo.Owner+"/"+ctx.GHRepo.Name, repo) &&
		gitspice.HasSpiceData(ctx.RepoPath) {
		return ctx.RepoPath
	}
	return ""
}

// stackState holds the stack folding state of a section.
type stackState struct {
	// folded holds the URLs of the roots whose stacks are folded.
	folded map[string]bool
	// hidden holds the PRs of folded stacks, which are not in Model.Prs.
	hidden []prrow.Data
}

// arrangeStacks groups stacked PRs under the root of their stack, sets their
// placement, and hides the members of folded stacks. firstPage tells if
// Model.Prs was just replaced by a new first page, which makes the hidden PRs
// stale.
func (m *Model) arrangeStacks(firstPage bool) {
	if !stacksEnabled(m.Ctx) {
		return
	}
	if m.stacks.folded == nil {
		m.stacks.folded = map[string]bool{}
	}

	all := m.Prs
	if !firstPage {
		all = append(all, m.stacks.hidden...)
	}

	items := make([]stack.Item, len(all))
	for i, pr := range all {
		items[i] = stack.Item{
			URL:     pr.Primary.Url,
			Repo:    pr.Primary.Repository.NameWithOwner,
			Head:    pr.Primary.HeadRefName,
			Base:    pr.Primary.BaseRefName,
			Sources: pr.Stack.Sources,
		}
	}

	order, placements := stack.Arrange(items)
	visible := make([]prrow.Data, 0, len(all))
	var hidden []prrow.Data
	for _, i := range order {
		pr := all[i]
		p := placements[i]
		pr.Stack.Placement = p
		pr.Stack.RootURL = ""
		pr.Stack.Folded = false
		if p != nil {
			pr.Stack.RootURL = all[p.Root].Primary.Url
			pr.Stack.Folded = m.stacks.folded[pr.Stack.RootURL]
		}
		if p != nil && !p.IsRoot && pr.Stack.Folded {
			hidden = append(hidden, pr)
			continue
		}
		visible = append(visible, pr)
	}
	m.Prs = visible
	m.stacks.hidden = hidden
	m.setStackRowDecorator()
}

// toggleStack folds or unfolds the stack of the current PR. When the current
// PR is a member of the stack, the cursor moves to the root first.
func (m *Model) toggleStack() {
	idx := m.Table.GetCurrItem()
	if !stacksEnabled(m.Ctx) || idx < 0 || idx >= len(m.Prs) {
		return
	}
	rootURL := m.Prs[idx].Stack.RootURL
	if rootURL == "" {
		return
	}

	for idx > 0 && m.Prs[idx].Primary.Url != rootURL {
		m.Table.PrevItem()
		idx--
	}
	m.stacks.folded[rootURL] = !m.stacks.folded[rootURL]
	m.arrangeStacks(false)
	m.Table.SetRows(m.BuildRows())
}

// setStackRowDecorator draws the stack rail through the separator line under
// each root and each member that is not the last of its stack.
func (m *Model) setStackRowDecorator() {
	if !m.Ctx.Config.Theme.Ui.Table.ShowSeparator {
		m.Table.SetRowDecorator(nil)
		return
	}

	stops := make([]string, len(m.Prs))
	for i, pr := range m.Prs {
		p := pr.Stack.Placement
		switch {
		case p == nil || pr.Stack.Folded:
		case p.IsRoot:
			stops[i] = "┬"
		case !p.Last:
			stops[i] = "┼"
		}
	}

	ctx := m.Ctx
	tbl := &m.Table
	m.Table.SetRowDecorator(func(rowId int, row string) string {
		if rowId < 0 || rowId >= len(stops) || stops[rowId] == "" {
			return row
		}
		lines := strings.Split(row, "\n")
		last := len(lines) - 1
		width := lipgloss.Width(lines[last])
		x := railColumn(tbl.Columns)
		if x >= width {
			return row
		}
		border := lipgloss.NewStyle().Foreground(ctx.Theme.FaintBorder)
		lines[last] = border.Render(strings.Repeat("─", x)) +
			lipgloss.NewStyle().Foreground(prrow.StackRailColor).Render(stops[rowId]) +
			border.Render(strings.Repeat("─", width-x-1))
		return strings.Join(lines, "\n")
	})
}

// railColumn returns the column of the rail in a row: the start of the text
// of the title column.
func railColumn(columns []table.Column) int {
	x := 0
	for _, c := range columns {
		if c.Hidden != nil && *c.Hidden {
			continue
		}
		if c.Grow != nil && *c.Grow {
			return x + 1
		}
		x += c.ComputedWidth
	}
	return x
}
