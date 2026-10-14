# Roadmap

Planned features of this fork, in the order to build them.

## 1. Hide unused features

Views are the `ViewType` enum in `internal/config/parser.go`; each view lives in
`internal/tui/components/<name>section`. Hide them through config, in the header
next to the repo selector.

## 2. Hide PRs

Hide PRs that I do not want in the list for a long time.

- A key hides the current PR. When it is the root of a stack, the stack is
  hidden with it.
- By default, a hidden PR comes back when it changes after it was hidden (its
  `updatedAt` is newer: new commits, comments, reviews, state changes). A
  second key hides it for good, until I unhide it.
- Hidden PRs are kept in a state file, like the done notifications
  (`internal/data/donestore.go`, `$XDG_STATE_HOME/gh-dash/`): URL → `updatedAt`
  when hidden, or a "for good" mark.
- The section shows how many PRs it hides, for example `3 hidden` next to the
  section title. A key shows the hidden PRs again (dimmed), and the hide key on
  a hidden PR unhides it.

## 3. "Needs my attention" signal

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

## 4. Vim mode in inputs

Inputs are stock Bubbles `textinput` / `textarea`:
`internal/tui/components/search/search.go`, `prompt/prompt.go`, `inputbox/inputbox.go`.
Build one modal wrapper (normal/insert, basic motions) and swap it in at those sites.

## Order

1. **Hide unused features** first. It is small and touches the header, next
   to the repo selector.
2. **Hide PRs** next, so the hidden count and the reveal key apply to the
   repo in scope.
3. **"Needs my attention"** after that. It needs the most new GraphQL fields
   and rules to refine.
4. **Vim mode** last. It is independent of the rest and is polish.
