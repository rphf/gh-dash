package prrow

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/stack"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newStackPR(p *stack.Placement, folded bool) *PullRequest {
	ctx := &context.ProgramContext{Theme: *theme.DefaultTheme, Config: &config.Config{}}
	ctx.Styles = context.InitStyles(ctx.Theme)
	return &PullRequest{
		Ctx: ctx,
		Data: &Data{
			Primary: &data.PullRequestData{Title: "add things"},
			Stack:   Stack{Placement: p, Folded: folded},
		},
	}
}

func TestWithStackExtended(t *testing.T) {
	render := func(pr *PullRequest) (string, string) {
		top, title := pr.withStackExtended(false, 30, "o/r #1 by @me", "add things")
		if pr.Data.Stack.Placement != nil {
			require.Equal(t, 30, ansi.StringWidth(title))
		}
		return strings.TrimRight(ansi.Strip(top), " "), strings.TrimRight(ansi.Strip(title), " ")
	}

	top, title := render(newStackPR(nil, false))
	require.Equal(t, "o/r #1 by @me", top)
	require.Equal(t, "add things", title)

	_, title = render(newStackPR(&stack.Placement{IsRoot: true, Members: 2}, false))
	require.Equal(t, "add things ▾ 2 stacked", title)

	_, title = render(newStackPR(&stack.Placement{IsRoot: true, Members: 2}, true))
	require.Equal(t, "add things ▸ 2 stacked", title)

	top, title = render(newStackPR(&stack.Placement{Position: 2}, false))
	require.Equal(t, "├ o/r #1 by @me", top)
	require.Equal(t, "│ 2 add things", title)

	top, title = render(newStackPR(&stack.Placement{Position: 3, Last: true}, false))
	require.Equal(t, "└ o/r #1 by @me", top)
	require.Equal(t, "  3 add things", title)
}

func TestLoneBadge(t *testing.T) {
	render := func(s stack.Sources) string {
		pr := newStackPR(nil, false)
		pr.Data.Stack.Sources = s
		_, title := pr.withStackExtended(false, 40, "#1", "add things")
		return strings.TrimRight(ansi.Strip(title), " ")
	}
	require.Equal(t, "add things", render(stack.Sources{}))
	require.Equal(t, "add things", render(stack.Sources{
		Native: &stack.Native{Size: 1, Position: 1},
	}))
	require.Equal(t, "add things 3 of 15 in stack", render(stack.Sources{
		Native: &stack.Native{Size: 15, Position: 3},
	}))
	require.Equal(t, "add things stacked on #12", render(stack.Sources{
		SpiceParentURL: "https://github.com/o/r/pull/12",
	}))
}

func TestStackNumberPadding(t *testing.T) {
	pr := newStackPR(&stack.Placement{Position: 5, Size: 11}, false)
	require.Equal(t, " 5 ", ansi.Strip(pr.renderStackNumber(false)))
	pr = newStackPR(&stack.Placement{Position: 10, Size: 11}, false)
	require.Equal(t, "10 ", ansi.Strip(pr.renderStackNumber(false)))
}
