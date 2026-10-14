package prssection

import (
	"slices"
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

// rowState holds what the section keeps to build its rows.
type rowState struct {
	// all holds every fetched PR, in fetch order. Model.Prs holds the PRs
	// shown, built from all.
	all []prrow.Data
	// folded holds the URLs of the roots whose stacks are folded.
	folded map[string]bool
	// showHidden shows the hidden PRs, dimmed.
	showHidden bool
	// hidden holds the URLs of the hidden PRs in all.
	hidden map[string]bool
	// roots holds the URL of the root of the stack of each stacked PR, by
	// URL, folded members included.
	roots map[string]string
}

// syncAll adds Model.Prs to the fetched PRs. firstPage tells if Model.Prs
// was just replaced by a new first page. PRs already fetched are replaced,
// so changes to Model.Prs are kept.
func (m *Model) syncAll(firstPage bool) {
	if firstPage {
		m.rows.all = slices.Clone(m.Prs)
		return
	}
	index := make(map[string]int, len(m.rows.all))
	for i, pr := range m.rows.all {
		index[pr.Primary.Url] = i
	}
	for _, pr := range m.Prs {
		if i, ok := index[pr.Primary.Url]; ok {
			m.rows.all[i] = pr
		} else {
			index[pr.Primary.Url] = len(m.rows.all)
			m.rows.all = append(m.rows.all, pr)
		}
	}
}

// arrangeStacks builds Model.Prs from the fetched PRs: it leaves out the
// hidden PRs, groups stacked PRs under the root of their stack, sets their
// placement, and leaves out the members of folded stacks. firstPage tells if
// Model.Prs was just replaced by a new first page.
func (m *Model) arrangeStacks(firstPage bool) {
	m.syncAll(firstPage)
	all := m.withoutHidden(slices.Clone(m.rows.all))
	if !stacksEnabled(m.Ctx) {
		m.Prs = m.sortByAttention(all, all)
		m.rows.roots = nil
		m.setRowDecorator()
		return
	}
	if m.rows.folded == nil {
		m.rows.folded = map[string]bool{}
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
	m.rows.roots = map[string]string{}
	visible := make([]prrow.Data, 0, len(all))
	arranged := make([]prrow.Data, 0, len(all))
	for _, i := range order {
		pr := all[i]
		p := placements[i]
		pr.Stack.Placement = p
		pr.Stack.RootURL = ""
		pr.Stack.Folded = false
		if p != nil {
			pr.Stack.RootURL = all[p.Root].Primary.Url
			pr.Stack.Folded = m.rows.folded[pr.Stack.RootURL]
			m.rows.roots[pr.Primary.Url] = pr.Stack.RootURL
		}
		arranged = append(arranged, pr)
		if p != nil && !p.IsRoot && pr.Stack.Folded {
			continue
		}
		visible = append(visible, pr)
	}
	m.Prs = m.sortByAttention(arranged, visible)
	m.setRowDecorator()
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
	m.rows.folded[rootURL] = !m.rows.folded[rootURL]
	m.arrangeStacks(false)
	m.Table.SetRows(m.BuildRows())
}

// setRowDecorator dims the hidden PRs, and draws the stack rail through the
// separator line under each root and each member that is not the last of its
// stack. Between members, the separator line shows only the rail.
func (m *Model) setRowDecorator() {
	separator := m.Ctx.Config.Theme.Ui.Table.ShowSeparator
	dimmed := make([]bool, len(m.Prs))
	stops := make([]string, len(m.Prs))
	for i, pr := range m.Prs {
		dimmed[i] = m.rows.hidden[pr.Primary.Url]
		p := pr.Stack.Placement
		switch {
		case p == nil || pr.Stack.Folded:
		case p.IsRoot:
			stops[i] = "┬"
		case !p.Last:
			stops[i] = "│"
		}
	}

	ctx := m.Ctx
	tbl := &m.Table
	m.Table.SetRowDecorator(func(rowId int, row string) string {
		if rowId < 0 || rowId >= len(stops) {
			return row
		}
		lines := strings.Split(row, "\n")
		last := len(lines) - 1
		if dimmed[rowId] {
			content := len(lines)
			if separator {
				content = last
			}
			dimLines(ctx, lines[:content], rowId == tbl.GetCurrItem())
		}
		if !separator || stops[rowId] == "" {
			return strings.Join(lines, "\n")
		}
		width := lipgloss.Width(lines[last])
		x := railColumn(tbl.Columns)
		if x >= width {
			return strings.Join(lines, "\n")
		}
		border := lipgloss.NewStyle().Foreground(ctx.Theme.FaintBorder)
		rail := lipgloss.NewStyle().Foreground(prrow.StackRailColor)
		if dimmed[rowId] {
			rail = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
		}
		fill := "─"
		if stops[rowId] == "│" {
			fill = " "
		}
		lines[last] = border.Render(strings.Repeat(fill, x)) +
			rail.Render(stops[rowId]) +
			border.Render(strings.Repeat(fill, width-x-1))
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
