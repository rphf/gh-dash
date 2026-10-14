package prrow

import (
	"fmt"
	"path"
	"strconv"
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
	rail := stackStyle(pr.Ctx, isSelected).Render("│ ")
	return rail, rail
}

// renderStackNumber renders the position of a stacked PR, such as "3 ". The
// numbers of members are padded to the width of the stack size so their
// titles line up.
func (pr *PullRequest) renderStackNumber(isSelected bool) string {
	if pr.Data.Stack.Placement == nil {
		return ""
	}
	p := pr.Data.Stack.Placement
	width := len(strconv.Itoa(max(p.Size, p.Position)))
	if p.IsRoot {
		width = 0
	}
	return stackStyle(pr.Ctx, isSelected).Bold(true).
		Render(fmt.Sprintf("%*d ", width, p.Position))
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

// renderLoneBadge renders, faint, the place in its stack of a stacked PR
// shown without the rest of its stack, such as " 3 of 15 in stack".
func (pr *PullRequest) renderLoneBadge(isSelected bool) string {
	if pr.Data.Stack.Placement != nil {
		return ""
	}
	var text string
	switch s := pr.Data.Stack.Sources; {
	case s.Native != nil && s.Native.Size > 1:
		text = fmt.Sprintf("%d of %d in stack", s.Native.Position, s.Native.Size)
	case s.SpiceParentURL != "":
		text = "stacked on #" + path.Base(s.SpiceParentURL)
	default:
		return ""
	}
	style := lipgloss.NewStyle().Foreground(pr.Ctx.Theme.FaintText)
	if isSelected {
		style = style.Background(pr.Ctx.Theme.SelectedBackground)
	}
	return style.Render(" " + text)
}

// withStackCompact adds the stack rail, number and badge to a compact title.
func (pr *PullRequest) withStackCompact(title string) string {
	lone := pr.renderLoneBadge(false)
	if pr.Data.Stack.Placement == nil && lone == "" {
		return title
	}
	rail, _ := pr.renderStackRails(false)
	return rail + pr.renderStackNumber(false) + title + pr.renderStackBadge(false) + lone
}

// withStackExtended adds the stack rail, number and badge to the two lines of
// an extended title, each rendered to width.
func (pr *PullRequest) withStackExtended(
	isSelected bool,
	width int,
	top, title string,
) (string, string) {
	lone := pr.renderLoneBadge(isSelected)
	if pr.Data.Stack.Placement == nil && lone == "" {
		return top, title
	}

	railTop, railTitle := pr.renderStackRails(isSelected)
	top = railTop + ansi.Truncate(top, width-lipgloss.Width(railTop), "")

	bg := lipgloss.NewStyle()
	if isSelected {
		bg = bg.Background(pr.Ctx.Theme.SelectedBackground)
	}
	number := pr.renderStackNumber(isSelected)
	badge := pr.renderStackBadge(isSelected) + lone
	avail := width - lipgloss.Width(railTitle) - lipgloss.Width(number) - lipgloss.Width(badge)
	text := bg.Foreground(pr.Ctx.Theme.PrimaryText).Bold(true).
		Render(ansi.Truncate(pr.Data.Primary.Title, max(avail, 0), constants.Ellipsis))
	line := railTitle + number + text + badge
	if pad := width - lipgloss.Width(line); pad > 0 {
		line += bg.Render(strings.Repeat(" ", pad))
	}
	return top, line
}
