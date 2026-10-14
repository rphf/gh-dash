package prssection

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
)

// refetch replaces the PRs of m with prs, as a new first page does.
func refetch(m *Model, prs ...prrow.Data) {
	m.ResetPageInfo()
	m.Prs = slices.Clone(prs)
	m.onPrsFetched(true)
	m.Table.SetRows(m.BuildRows())
}

func currNumber(m *Model) int {
	return m.Prs[m.Table.GetCurrItem()].Primary.Number
}

func TestRefreshKeepsRows(t *testing.T) {
	m := newStackTestModel(t)
	m.ResetPageInfo()
	m.SetIsLoading(true)
	require.False(t, m.Table.IsLoading(), "rows stay while refreshing")
}

func TestRefreshKeepsCursor(t *testing.T) {
	m := newStackTestModel(t)
	all := slices.Clone(m.rows.all) // 1 <- 2 <- 3 stacked, then 4
	m.Table.LastItem()
	require.Equal(t, 4, currNumber(m))

	refetch(m, all[3], all[0], all[1], all[2])
	require.Equal(t, []int{4, 1, 2, 3}, numbers(m))
	require.Equal(t, 4, currNumber(m), "cursor follows the PR up")

	newPR := prrow.Data{Primary: &data.PullRequestData{
		Number:     5,
		Url:        "https://github.com/o/r/pull/e",
		Repository: data.Repository{NameWithOwner: "o/r"},
	}}
	m.Table.LastItem()
	require.Equal(t, 3, currNumber(m))
	refetch(m, newPR, all[3], all[0], all[1], all[2])
	require.Equal(t, []int{5, 4, 1, 2, 3}, numbers(m))
	require.Equal(t, 3, currNumber(m), "cursor moves past the old last row")

	refetch(m, all[3])
	require.Equal(t, 0, m.Table.GetCurrItem(), "cursor stays in the rows")
}
