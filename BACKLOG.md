# Roadmap

Planned features of this fork, in the order to build them.

## 1. Vim mode in inputs

Inputs are stock Bubbles `textinput` / `textarea`:
`internal/tui/components/search/search.go`, `prompt/prompt.go`, `inputbox/inputbox.go`.
Build one modal wrapper (normal/insert, basic motions) and swap it in at those sites.

## 2. Hide unused features

Views are the `ViewType` enum in `internal/config/parser.go`; each view lives in
`internal/tui/components/<name>section`. Hide them through config, in the header
next to the repo selector.

## Order

1. **Vim mode** first. It is independent of the rest.
2. **Hide unused features** next. It is small and touches only the header.
