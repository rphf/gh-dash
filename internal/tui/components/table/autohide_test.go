package table

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func isHidden(c Column) bool {
	return c.Hidden != nil && *c.Hidden
}

func TestHideEmptyColumns(t *testing.T) {
	ctx := context.ProgramContext{Theme: *theme.DefaultTheme, Config: &config.Config{
		Theme: &config.ThemeConfig{},
	}}
	ctx.Styles = context.InitStyles(ctx.Theme)
	hidden, grow := true, true
	m := NewModel(ctx, constants.Dimensions{Width: 80, Height: 10}, time.Time{},
		time.Time{}, []Column{
			{Title: "a"},
			{Title: "title", Grow: &grow},
			{Title: "b"},
			{Title: "c", Hidden: &hidden},
		}, nil, "row", nil, "", false)

	m.HideEmptyColumns(nil)
	require.False(t, isHidden(m.Columns[0]))

	m.HideEmptyColumns([]Row{{"x", "", "  ", "z"}, {"y", "", "", ""}})
	require.False(t, isHidden(m.Columns[0]))
	require.False(t, isHidden(m.Columns[1]), "grow column stays")
	require.True(t, isHidden(m.Columns[2]))
	require.True(t, isHidden(m.Columns[3]))

	m.HideEmptyColumns([]Row{{"x", "", "b", "z"}})
	require.False(t, isHidden(m.Columns[2]))
	require.True(t, isHidden(m.Columns[3]), "config hidden column stays hidden")
}
