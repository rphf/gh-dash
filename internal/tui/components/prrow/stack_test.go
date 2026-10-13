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
	require.Equal(t, "│ ② add things", title)

	top, title = render(newStackPR(&stack.Placement{Position: 3, Last: true}, false))
	require.Equal(t, "└ o/r #1 by @me", top)
	require.Equal(t, "  ③ add things", title)
}

func TestCircledNumber(t *testing.T) {
	require.Equal(t, "①", circledNumber(1))
	require.Equal(t, "⑳", circledNumber(20))
	require.Equal(t, "㉑", circledNumber(21))
	require.Equal(t, "㊿", circledNumber(50))
	require.Equal(t, "(51)", circledNumber(51))
}
