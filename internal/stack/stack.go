// Package stack groups pull requests into stacks and orders them.
//
// A PR's parent (the PR below it) comes from the first source that knows it:
//  1. GitHub native stacks (stack number and position).
//  2. git-spice in a local clone (the PR of the branch below).
//  3. The base branch of the PR, matched to the head branch of another PR.
package stack

import "sort"

// Native is the GitHub native stack data of a PR.
type Native struct {
	Number   int
	Size     int
	Position int
}

// Sources holds what each source knows about a PR.
type Sources struct {
	Native *Native
	// SpiceParentURL is the URL of the PR below this one in git-spice.
	SpiceParentURL string
}

// Placement is the position of a PR in its stack. The root of a stack is
// its lowest PR. It is shown at the top level, and the other PRs of the stack
// are its members, shown under it.
type Placement struct {
	// Position is 1-based. The PR at the bottom of the stack has position 1.
	Position int
	Size     int
	// Root is the index, in the input items, of the root of the stack.
	Root   int
	IsRoot bool
	// Members is the number of PRs under the root. It is set on the root only.
	Members int
	// Last tells if the PR is the last row of its stack.
	Last bool
}

// Item is the input for one PR.
type Item struct {
	URL     string
	Repo    string
	Head    string
	Base    string
	Sources Sources
}

type nativeKey struct {
	repo   string
	number int
}

type branchKey struct {
	repo   string
	branch string
}

// Arrange returns the display order of items, as indexes into items, and the
// placement of each item. A placement is nil when the item is not in a stack.
// A PR alone in the input is not in a stack, even when its native stack has
// more PRs. Each stack is shown as one contiguous group, from bottom to top,
// at the place of its first item in the input order.
func Arrange(items []Item) ([]int, []*Placement) {
	n := len(items)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = -1
	}

	byURL := make(map[string]int, n)
	byHead := make(map[branchKey]int, n)
	byNative := make(map[nativeKey][]int)
	for i, it := range items {
		if _, ok := byURL[it.URL]; !ok && it.URL != "" {
			byURL[it.URL] = i
		}
		k := branchKey{it.Repo, it.Head}
		if _, ok := byHead[k]; !ok && it.Head != "" {
			byHead[k] = i
		}
		if nat := it.Sources.Native; nat != nil {
			nk := nativeKey{it.Repo, nat.Number}
			byNative[nk] = append(byNative[nk], i)
		}
	}

	createsCycle := func(child, candidate int) bool {
		for p := candidate; p != -1; p = parent[p] {
			if p == child {
				return true
			}
		}
		return false
	}

	for i, it := range items {
		candidate := -1
		if nat := it.Sources.Native; nat != nil {
			bestPos := 0
			for _, j := range byNative[nativeKey{it.Repo, nat.Number}] {
				pos := items[j].Sources.Native.Position
				if j != i && pos < nat.Position && pos > bestPos {
					candidate, bestPos = j, pos
				}
			}
		}
		if candidate == -1 && it.Sources.SpiceParentURL != "" {
			if j, ok := byURL[it.Sources.SpiceParentURL]; ok && j != i {
				candidate = j
			}
		}
		if candidate == -1 && it.Base != "" {
			if j, ok := byHead[branchKey{it.Repo, it.Base}]; ok && j != i {
				candidate = j
			}
		}
		if candidate != -1 && !createsCycle(i, candidate) {
			parent[i] = candidate
		}
	}

	children := make([][]int, n)
	for i, p := range parent {
		if p != -1 {
			children[p] = append(children[p], i)
		}
	}
	for _, c := range children {
		sort.SliceStable(c, func(a, b int) bool {
			return nativePosition(items[c[a]]) < nativePosition(items[c[b]])
		})
	}

	order := make([]int, 0, n)
	placements := make([]*Placement, n)
	emitted := make([]bool, n)
	for i := range items {
		root := i
		for parent[root] != -1 {
			root = parent[root]
		}
		if emitted[root] {
			continue
		}

		var group []int
		positions := map[int]int{}
		var walk func(node, pos int)
		walk = func(node, pos int) {
			if nat := items[node].Sources.Native; nat != nil {
				pos = nat.Position
			}
			emitted[node] = true
			positions[node] = pos
			group = append(group, node)
			for _, c := range children[node] {
				walk(c, pos+1)
			}
		}
		walk(root, 1)
		order = append(order, group...)

		size := 0
		for _, node := range group {
			size = max(size, positions[node])
			if nat := items[node].Sources.Native; nat != nil {
				size = max(size, nat.Size)
			}
		}
		if len(group) == 1 {
			continue
		}
		for gi, node := range group {
			placements[node] = &Placement{
				Position: positions[node],
				Size:     size,
				Root:     root,
				IsRoot:   gi == 0,
				Last:     gi == len(group)-1,
			}
		}
		placements[root].Members = len(group) - 1
	}

	return order, placements
}

func nativePosition(it Item) int {
	if it.Sources.Native != nil {
		return it.Sources.Native.Position
	}
	return 0
}
