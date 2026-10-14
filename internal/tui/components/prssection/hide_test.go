package prssection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHideStackRoot(t *testing.T) {
	m := newStackTestModel(t)
	m.toggleHidden()
	require.Equal(t, []int{4}, numbers(m))
	require.Len(t, m.rows.hidden, 3)
	require.Contains(t, m.withHiddenCount("PR 1/4"), "3 hidden (. to show)")
	require.Equal(t, 0, m.Table.GetCurrItem())

	m.toggleShowHidden()
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))
	require.True(t, m.rows.hidden[m.Prs[0].Primary.Url])

	m.toggleHidden()
	require.Empty(t, m.rows.hidden)
	require.Equal(t, "PR 1/4", m.withHiddenCount("PR 1/4"))
	m.toggleShowHidden()
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))
}

func TestHiddenPRComesBackWhenItChanges(t *testing.T) {
	m := newStackTestModel(t)
	m.Table.LastItem()
	m.toggleHidden()
	require.Equal(t, []int{1, 2, 3}, numbers(m))
	require.Equal(t, 2, m.Table.GetCurrItem())

	x := m.rows.all[3]
	x.Primary.UpdatedAt = time.Now()
	m.Prs = append(m.Prs, x)
	m.arrangeStacks(false)
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))
}
