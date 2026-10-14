package prrow

import (
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/attention"
)

// withAttention adds the attention status of the PR to the top line of an
// extended row, such as " · your move: changes requested".
func (pr *PullRequest) withAttention(baseStyle lipgloss.Style, top string) string {
	status := pr.Data.Attention
	label := status.Label()
	if label == "" {
		return top
	}
	style := baseStyle.Foreground(pr.Ctx.Theme.WarningText)
	if status.Bucket == attention.ReadyToShip {
		style = baseStyle.Foreground(pr.Ctx.Theme.SuccessText)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, top, baseStyle.Render(" · "), style.Render(label))
}
