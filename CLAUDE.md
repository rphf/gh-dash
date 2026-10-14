# gh-dash — personal fork

Personal flavor of [dlvhdr/gh-dash](https://github.com/dlvhdr/gh-dash).
The only remote is `origin` = `rphf/gh-dash`, and `gh` defaults to `rphf/gh-dash`.
Do not add a remote for `dlvhdr/gh-dash`.
Planned features: `BACKLOG.md`.

## Never upstream

Upstream has a strict no-AI contribution policy (`AI_POLICY.md`, `CONTRIBUTING.md`).
Code in this fork is AI-assisted, so it stays here. Do not open issues, PRs or
discussions on `dlvhdr/gh-dash`.

## Keep upstream rebases cheap

- `main` holds the fork's changes as a small set of commits on top of upstream `main`.
  Sync by fetching upstream by URL, without a remote:
  `git fetch https://github.com/dlvhdr/gh-dash.git main && git rebase FETCH_HEAD`.
- Put new code in new files and new packages. Touch upstream files only to add a
  hook point (one call, one field, one case), never to restructure them.
- Disable features through config or by not wiring them in. Do not delete upstream code.
- Gate new behavior behind config keys with defaults that match upstream behavior.
- Do not reformat, rename or move upstream code. Do not bump dependencies unless a
  feature needs it.
- Keep one feature per commit (or a few), with Conventional Commits messages, so a
  conflicting commit is easy to find and fix during a rebase.
- Prefer extending GraphQL query structs with new fields over rewriting queries.

## Build and run

- Build: `go build .` (produces `./gh-dash`).
- Install as the `gh dash` extension: `gh ext remove dash; gh ext install .`
  (symlink to `./gh-dash`, so each `go build` updates it).
- Tests: `go test ./...`. Lint: `golangci-lint run --config=.golangci.yml`.
  Format: `gofumpt -w`.
- Debug logs: see the `debug` and `logs` tasks in `Taskfile.yaml`.
