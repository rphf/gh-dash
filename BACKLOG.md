# Roadmap

Planned features of this fork, in the order to build them.

## 1. "Needs my attention" signal

For PRs I am involved in. Draft criteria, to refine:

| Bucket | Rule |
|---|---|
| My move (reviewer) | Review requested from me, not draft, no review from me since the last push |
| My move (author) | Changes requested, new comments since my last reply, CI failing, or merge conflict |
| Ready to ship (author) | Approved, CI green, mergeable |
| Waiting on others | Everything else |
| Ignored | Drafts I am not the author of |

Most inputs are already in `internal/data/prapi.go` (`ReviewDecision`,
`ReviewRequests`, `IsDraft`, `Mergeable`, `StatusCheckRollup`). "Who acted last"
needs extra fields (latest reviews and comments with author and timestamp).

## 2. Vim mode in inputs

Inputs are stock Bubbles `textinput` / `textarea`:
`internal/tui/components/search/search.go`, `prompt/prompt.go`, `inputbox/inputbox.go`.
Build one modal wrapper (normal/insert, basic motions) and swap it in at those sites.

## 3. Hide unused features

Views are the `ViewType` enum in `internal/config/parser.go`; each view lives in
`internal/tui/components/<name>section`. Hide them through config, in the header
next to the repo selector.

## Order

1. **"Needs my attention"** first. It needs the most new GraphQL fields and
   rules to refine.
2. **Vim mode** next. It is independent of the rest and is polish.
3. **Hide unused features** last. It is small and touches only the header.
