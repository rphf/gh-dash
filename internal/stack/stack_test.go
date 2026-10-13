package stack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func url(n string) string { return "https://github.com/o/r/pull/" + n }

func TestArrangeBaseChain(t *testing.T) {
	items := []Item{
		{URL: url("3"), Repo: "o/r", Head: "c", Base: "b"},
		{URL: url("9"), Repo: "o/r", Head: "x", Base: "main"},
		{URL: url("1"), Repo: "o/r", Head: "a", Base: "main"},
		{URL: url("2"), Repo: "o/r", Head: "b", Base: "a"},
	}
	order, pl := Arrange(items)
	require.Equal(t, []int{2, 3, 0, 1}, order)
	require.Equal(t, &Placement{Position: 1, Size: 3, Root: 2, IsRoot: true, Members: 2}, pl[2])
	require.Equal(t, &Placement{Position: 2, Size: 3, Root: 2}, pl[3])
	require.Equal(t, &Placement{Position: 3, Size: 3, Root: 2, Last: true}, pl[0])
	require.Nil(t, pl[1])
}

func TestArrangeBaseChainOtherRepo(t *testing.T) {
	items := []Item{
		{URL: url("1"), Repo: "o/r", Head: "a", Base: "main"},
		{URL: "https://github.com/o/s/pull/2", Repo: "o/s", Head: "b", Base: "a"},
	}
	order, pl := Arrange(items)
	require.Equal(t, []int{0, 1}, order)
	require.Nil(t, pl[0])
	require.Nil(t, pl[1])
}

func TestArrangeNative(t *testing.T) {
	items := []Item{
		{
			URL: url("15"), Repo: "o/r", Head: "p", Base: "q",
			Sources: Sources{Native: &Native{Number: 7, Size: 15, Position: 15}},
		},
		{
			URL: url("4"), Repo: "o/r", Head: "d", Base: "zzz",
			Sources: Sources{Native: &Native{Number: 7, Size: 15, Position: 2}},
		},
		{
			URL: url("5"), Repo: "o/r", Head: "solo", Base: "main",
			Sources: Sources{Native: &Native{Number: 8, Size: 3, Position: 2}},
		},
	}
	order, pl := Arrange(items)
	require.Equal(t, []int{1, 0, 2}, order)
	require.Equal(t, &Placement{Position: 2, Size: 15, Root: 1, IsRoot: true, Members: 1}, pl[1])
	require.Equal(t, &Placement{Position: 15, Size: 15, Root: 1, Last: true}, pl[0])
	require.Nil(t, pl[2])
}

func TestArrangeSpiceWinsOverBase(t *testing.T) {
	items := []Item{
		{URL: url("1"), Repo: "o/r", Head: "a", Base: "main"},
		{URL: url("2"), Repo: "o/r", Head: "b", Base: "main"},
		{
			URL: url("3"), Repo: "o/r", Head: "c", Base: "a",
			Sources: Sources{SpiceParentURL: url("2")},
		},
	}
	order, pl := Arrange(items)
	require.Equal(t, []int{0, 1, 2}, order)
	require.Nil(t, pl[0])
	require.True(t, pl[1].IsRoot)
	require.Equal(t, 1, pl[2].Root)
}

func TestArrangeCycle(t *testing.T) {
	items := []Item{
		{URL: url("1"), Repo: "o/r", Head: "a", Base: "b"},
		{URL: url("2"), Repo: "o/r", Head: "b", Base: "a"},
	}
	order, pl := Arrange(items)
	require.ElementsMatch(t, []int{0, 1}, order)
	require.NotNil(t, pl[0])
	require.NotNil(t, pl[1])
}

func TestArrangeTree(t *testing.T) {
	items := []Item{
		{URL: url("1"), Repo: "o/r", Head: "a", Base: "main"},
		{URL: url("2"), Repo: "o/r", Head: "b1", Base: "a"},
		{URL: url("3"), Repo: "o/r", Head: "b2", Base: "a"},
		{URL: url("4"), Repo: "o/r", Head: "c", Base: "b1"},
	}
	order, pl := Arrange(items)
	require.Equal(t, []int{0, 1, 3, 2}, order)
	require.Equal(t, 3, pl[0].Members)
	require.Equal(t, 3, pl[3].Position)
	require.Equal(t, 2, pl[2].Position)
	require.True(t, pl[2].Last)
}
