package prrow

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// leanTopLine returns the top line of an extended row with the repo shown
// as RowRepoName and without the author when it is the viewer. It returns
// top when the repo scope is off.
func (pr *PullRequest) leanTopLine(baseStyle lipgloss.Style, author, top string) string {
	if !pr.Ctx.RepoScoped() {
		return top
	}
	line := fmt.Sprintf("#%d", pr.Data.Primary.Number)
	if repo := pr.Ctx.RowRepoName(pr.Data.Primary.Repository.NameWithOwner); repo != "" {
		line = repo + " " + line
	}
	if !pr.isViewerAuthor() {
		line = lipgloss.JoinHorizontal(lipgloss.Top, line, baseStyle.Render(" by "), author)
	}
	return line
}

// isViewerAuthor tells if the repo scope is on and the viewer is the author.
func (pr *PullRequest) isViewerAuthor() bool {
	return pr.Ctx.RepoScoped() && pr.Data.Primary != nil &&
		pr.Ctx.IsViewer(pr.Data.Primary.Author.Login)
}

// isDefaultBase tells if the repo scope is on and the PR merges into the
// default branch of its repo.
func (pr *PullRequest) isDefaultBase() bool {
	if !pr.Ctx.RepoScoped() || pr.Data.Primary == nil {
		return false
	}
	def := pr.Data.Primary.Repository.DefaultBranchRef.Name
	return def != "" && def == pr.Data.Primary.BaseRefName
}

func (pr *PullRequest) renderScopedRepoName() string {
	return pr.getTextStyle().Foreground(pr.Ctx.Theme.FaintText).
		Render(pr.Ctx.RowRepoName(pr.Data.Primary.Repository.NameWithOwner))
}
