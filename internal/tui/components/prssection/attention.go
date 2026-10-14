package prssection

import (
	"slices"
	"sync"

	"charm.land/log/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/attention"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// loadRowExtras fills the stack sources and the attention status of prs. It
// runs in the fetch command.
func loadRowExtras(ctx *context.ProgramContext, prs []prrow.Data) {
	var wg sync.WaitGroup
	wg.Go(func() { loadStackSources(ctx, prs) })
	wg.Go(func() { loadAttention(ctx, prs) })
	wg.Wait()
}

func attentionEnabled(ctx *context.ProgramContext) bool {
	return ctx != nil && ctx.Config != nil && ctx.Config.Attention.IsEnabled()
}

// loadAttention fills the attention status of prs.
func loadAttention(ctx *context.ProgramContext, prs []prrow.Data) {
	if !attentionEnabled(ctx) || len(prs) == 0 {
		return
	}
	urls := make([]string, 0, len(prs))
	for _, pr := range prs {
		urls = append(urls, pr.Primary.Url)
	}
	viewer, inputs, err := data.FetchAttentionInputs(urls)
	if err != nil {
		log.Warn("failed to fetch PR attention inputs", "err", err)
		return
	}
	if viewer == "" {
		viewer = ctx.User
	}
	for i := range prs {
		pr := prs[i].Primary
		in := inputs[pr.Url]
		in.Viewer = viewer
		in.MyAuthors = ctx.Config.Attention.MyAuthors
		in.Author = pr.Author.Login
		in.IsDraft = pr.IsDraft
		in.ReviewDecision = pr.ReviewDecision
		in.Mergeable = pr.Mergeable
		if len(pr.Commits.Nodes) > 0 {
			in.CIState = string(pr.Commits.Nodes[0].Commit.StatusCheckRollup.State)
		}
		prs[i].Attention = attention.Compute(in)
	}
}

// groupKey returns the URL of the root of the stack of pr, or its own URL
// when it is not in a stack.
func groupKey(pr prrow.Data) string {
	if pr.Stack.RootURL != "" {
		return pr.Stack.RootURL
	}
	return pr.Primary.Url
}

// sortByAttention puts first the PRs that need the viewer, then the PRs
// ready to ship, when the config asks for it. A stack moves as one group,
// ranked by its PR that needs the viewer most, among all, folded PRs
// included. The order is stable otherwise.
func (m *Model) sortByAttention(all, shown []prrow.Data) []prrow.Data {
	if m.Ctx == nil || m.Ctx.Config == nil || !m.Ctx.Config.Attention.IsSortFirst() {
		return shown
	}
	rank := map[string]int{}
	for _, pr := range all {
		k, r := groupKey(pr), pr.Attention.Bucket.Rank()
		if cur, ok := rank[k]; !ok || r < cur {
			rank[k] = r
		}
	}

	var groups [][]prrow.Data
	for _, pr := range shown {
		if n := len(groups); n > 0 && groupKey(groups[n-1][0]) == groupKey(pr) {
			groups[n-1] = append(groups[n-1], pr)
			continue
		}
		groups = append(groups, []prrow.Data{pr})
	}
	slices.SortStableFunc(groups, func(a, b []prrow.Data) int {
		return rank[groupKey(a[0])] - rank[groupKey(b[0])]
	})
	return slices.Concat(groups...)
}
