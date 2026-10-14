package prrow

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newLeanPR(scope string, author, base string) *PullRequest {
	ctx := &context.ProgramContext{Theme: *theme.DefaultTheme, Config: &config.Config{}, User: "me"}
	ctx.Styles = context.InitStyles(ctx.Theme)
	ctx.InitRepoScope()
	ctx.SetScopeRepo(scope)
	ctx.AddKnownRepos("o/app")
	pr := &data.PullRequestData{Number: 7, BaseRefName: base}
	pr.Author.Login = author
	pr.Repository.NameWithOwner = "o/app"
	pr.Repository.DefaultBranchRef.Name = "main"
	return &PullRequest{Ctx: ctx, Data: &Data{Primary: pr}}
}

func TestLeanTopLine(t *testing.T) {
	top := func(pr *PullRequest) string {
		return ansi.Strip(
			pr.leanTopLine(lipgloss.NewStyle(), "@"+pr.Data.Primary.Author.Login, "full"),
		)
	}
	require.Equal(t, "#7", top(newLeanPR("o/app", "me", "main")))
	require.Equal(t, "#7 by @bob", top(newLeanPR("o/app", "bob", "main")))
	require.Equal(t, "app #7", top(newLeanPR("", "me", "main")))

	pr := newLeanPR("o/app", "me", "main")
	pr.Ctx.Scope = nil
	require.Equal(t, "full", top(pr))
}

func TestIsDefaultBase(t *testing.T) {
	require.True(t, newLeanPR("o/app", "me", "main").isDefaultBase())
	require.False(t, newLeanPR("o/app", "me", "feature").isDefaultBase())
}
