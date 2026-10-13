package prrow

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/charmbracelet/x/ansi"

	"github.com/dlvhdr/gh-dash/v4/internal/stack"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// StackRailColor is the color of the rail that joins the PRs of a stack.
var StackRailColor = compat.AdaptiveColor{
	Light: lipgloss.ANSIColor(6),
	Dark:  lipgloss.ANSIColor(14),
}

// Stack holds the stack data of a PR row.
type Stack struct {
	Sources   stack.Sources
	Placement *stack.Placement
	// RootURL is the URL of the root of the stack of the PR.
	RootURL string
	// Folded tells if the members of the stack are hidden. It is set on the
	// root only.
	Folded bool
}

func stackStyle(ctx *context.ProgramContext, isSelected bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(StackRailColor)
	if isSelected {
		s = s.Background(ctx.Theme.SelectedBackground)
	}
	return s
}

func (pr *PullRequest) isStackMember() bool {
	p := pr.Data.Stack.Placement
	return p != nil && !p.IsRoot
}

// renderStackRails renders the rail left of the first and the second line of
// a stack member, or empty strings for other PRs.
func (pr *PullRequest) renderStackRails(isSelected bool) (string, string) {
	if !pr.isStackMember() {
		return "", ""
	}
	style := stackStyle(pr.Ctx, isSelected)
	if pr.Data.Stack.Placement.Last {
		return style.Render("└ "), style.Render("  ")
	}
	return style.Render("├ "), style.Render("│ ")
}

// renderStackNumber renders the position of a stack member, such as "③ ".
func (pr *PullRequest) renderStackNumber(isSelected bool) string {
	if !pr.isStackMember() {
		return ""
	}
	return stackStyle(pr.Ctx, isSelected).Bold(true).
		Render(circledNumber(pr.Data.Stack.Placement.Position) + " ")
}

// renderStackBadge renders the number of PRs stacked on a root, and if they
// are folded.
func (pr *PullRequest) renderStackBadge(isSelected bool) string {
	p := pr.Data.Stack.Placement
	if p == nil || !p.IsRoot {
		return ""
	}
	arrow := "▾"
	if pr.Data.Stack.Folded {
		arrow = "▸"
	}
	return stackStyle(pr.Ctx, isSelected).
		Render(fmt.Sprintf(" %s %d stacked", arrow, p.Members))
}

// circledNumber returns n as a circled digit, such as ③, for 1 to 50.
func circledNumber(n int) string {
	switch {
	case n >= 1 && n <= 20:
		return string(rune('①' + n - 1))
	case n >= 21 && n <= 35:
		return string(rune('㉑' + n - 21))
	case n >= 36 && n <= 50:
		return string(rune('㊱' + n - 36))
	default:
		return fmt.Sprintf("(%d)", n)
	}
}

// withStackCompact adds the stack rail, number and badge to a compact title.
func (pr *PullRequest) withStackCompact(title string) string {
	if pr.Data.Stack.Placement == nil {
		return title
	}
	rail, _ := pr.renderStackRails(false)
	return rail + pr.renderStackNumber(false) + title + pr.renderStackBadge(false)
}

// withStackExtended adds the stack rail, number and badge to the two lines of
// an extended title, each rendered to width.
func (pr *PullRequest) withStackExtended(
	isSelected bool,
	width int,
	top, title string,
) (string, string) {
	if pr.Data.Stack.Placement == nil {
		return top, title
	}

	railTop, railTitle := pr.renderStackRails(isSelected)
	top = railTop + ansi.Truncate(top, width-lipgloss.Width(railTop), "")

	bg := lipgloss.NewStyle()
	if isSelected {
		bg = bg.Background(pr.Ctx.Theme.SelectedBackground)
	}
	number := pr.renderStackNumber(isSelected)
	badge := pr.renderStackBadge(isSelected)
	avail := width - lipgloss.Width(railTitle) - lipgloss.Width(number) - lipgloss.Width(badge)
	text := bg.Foreground(pr.Ctx.Theme.PrimaryText).Bold(true).
		Render(ansi.Truncate(pr.Data.Primary.Title, max(avail, 0), constants.Ellipsis))
	line := railTitle + number + text + badge
	if pad := width - lipgloss.Width(line); pad > 0 {
		line += bg.Render(strings.Repeat(" ", pad))
	}
	return top, line
}
