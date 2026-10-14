package prssection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newStackTestModel(t *testing.T) *Model {
	t.Helper()
	t.Cleanup(data.SetHiddenPRStoreForTesting(""))
	ctx := &context.ProgramContext{
		Theme: *theme.DefaultTheme,
		Config: &config.Config{
			Theme: &config.ThemeConfig{Ui: config.UIThemeConfig{
				Table: config.TableUIThemeConfig{ShowSeparator: true},
			}},
		},
	}
	ctx.Styles = context.InitStyles(ctx.Theme)
	m := NewModel(0, ctx, config.PrsSectionConfig{}, time.Now(), time.Now())

	pr := func(n int, head, base string) prrow.Data {
		return prrow.Data{Primary: &data.PullRequestData{
			Number:      n,
			Url:         "https://github.com/o/r/pull/" + head,
			Repository:  data.Repository{NameWithOwner: "o/r"},
			HeadRefName: head,
			BaseRefName: base,
		}}
	}
	m.Prs = []prrow.Data{
		pr(1, "a", "main"),
		pr(2, "b", "a"),
		pr(3, "c", "b"),
		pr(4, "x", "main"),
	}
	m.arrangeStacks(true)
	m.Table.SetRows(m.BuildRows())
	return &m
}

func numbers(m *Model) []int {
	var res []int
	for _, pr := range m.Prs {
		res = append(res, pr.Primary.Number)
	}
	return res
}

func TestToggleStackFromMember(t *testing.T) {
	m := newStackTestModel(t)
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))

	m.Table.NextItem()
	m.Table.NextItem()
	m.toggleStack()
	require.Equal(t, []int{1, 4}, numbers(m))
	require.Equal(t, 0, m.Table.GetCurrItem())
	require.True(t, m.Prs[0].Stack.Folded)

	m.toggleStack()
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))
	require.Equal(t, 0, m.Table.GetCurrItem())
}

func TestToggleStackKeepsFoldOnNextPage(t *testing.T) {
	m := newStackTestModel(t)
	m.toggleStack()
	require.Equal(t, []int{1, 4}, numbers(m))

	m.Prs = append(m.Prs, prrow.Data{Primary: &data.PullRequestData{
		Number:      5,
		Url:         "https://github.com/o/r/pull/d",
		Repository:  data.Repository{NameWithOwner: "o/r"},
		HeadRefName: "d",
		BaseRefName: "c",
	}})
	m.arrangeStacks(false)
	require.Equal(t, []int{1, 4}, numbers(m))
	require.Equal(t, 3, m.Prs[0].Stack.Placement.Members)
}
