package prssection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/attention"
)

func setBucket(m *Model, number int, b attention.Bucket) {
	for i := range m.rows.all {
		if m.rows.all[i].Primary.Number == number {
			m.rows.all[i].Attention = attention.Status{Bucket: b}
		}
	}
}

func TestSortByAttention(t *testing.T) {
	m := newStackTestModel(t)
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))

	setBucket(m, 4, attention.ReadyToShip)
	m.Prs = nil
	m.arrangeStacks(false)
	require.Equal(t, []int{4, 1, 2, 3}, numbers(m), "ready to ship before waiting")

	setBucket(m, 3, attention.MyMove)
	m.Prs = nil
	m.arrangeStacks(false)
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m), "the stack moves as one group")

	m.toggleStack()
	require.Equal(t, []int{1, 4}, numbers(m), "folded members still rank their stack")
}

func TestSortByAttentionOff(t *testing.T) {
	m := newStackTestModel(t)
	off := false
	m.Ctx.Config.Attention.SortFirst = &off
	setBucket(m, 4, attention.MyMove)
	m.Prs = nil
	m.arrangeStacks(false)
	require.Equal(t, []int{1, 2, 3, 4}, numbers(m))
}
