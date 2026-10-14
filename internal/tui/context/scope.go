package context

import (
	"slices"
	"strings"
	"sync"
)

// RepoScope is the repo whose PRs, issues and notifications are shown. An
// empty repo means all repos.
type RepoScope struct {
	mu   sync.RWMutex
	repo string
	// last is the last repo in scope, the target of the toggle from all repos.
	last string
	// known holds the repos seen in fetched rows, by lower-case name.
	known map[string]string
}

// InitRepoScope sets the scope at launch: the repo gh-dash runs in, when
// smart filtering at launch is on, or all repos.
func (ctx *ProgramContext) InitRepoScope() {
	if ctx.Config == nil || !ctx.Config.RepoScope.IsEnabled() {
		ctx.Scope = nil
		return
	}
	ctx.Scope = &RepoScope{known: map[string]string{}}
	if ctx.HasGHRepo() {
		home := ctx.GHRepo.Owner + "/" + ctx.GHRepo.Name
		ctx.Scope.last = home
		if ctx.Config.SmartFilteringAtLaunch {
			ctx.Scope.repo = home
		}
	}
}

// RepoScoped tells if the repo scope is on.
func (ctx *ProgramContext) RepoScoped() bool {
	return ctx != nil && ctx.Scope != nil
}

// ScopeRepo returns the repo in scope, or an empty string for all repos.
func (ctx *ProgramContext) ScopeRepo() string {
	if !ctx.RepoScoped() {
		return ""
	}
	ctx.Scope.mu.RLock()
	defer ctx.Scope.mu.RUnlock()
	return ctx.Scope.repo
}

// HomeRepo returns the repo gh-dash runs in, or an empty string.
func (ctx *ProgramContext) HomeRepo() string {
	if !ctx.HasGHRepo() {
		return ""
	}
	return ctx.GHRepo.Owner + "/" + ctx.GHRepo.Name
}

// SetScopeRepo sets the repo in scope. An empty repo means all repos.
func (ctx *ProgramContext) SetScopeRepo(repo string) {
	if !ctx.RepoScoped() {
		return
	}
	ctx.Scope.mu.Lock()
	defer ctx.Scope.mu.Unlock()
	ctx.Scope.repo = repo
	if repo != "" {
		ctx.Scope.last = repo
	}
}

// ToggledScopeRepo returns the scope the toggle goes to: all repos from a
// repo, or the last repo in scope from all repos.
func (ctx *ProgramContext) ToggledScopeRepo() string {
	if !ctx.RepoScoped() {
		return ""
	}
	ctx.Scope.mu.RLock()
	defer ctx.Scope.mu.RUnlock()
	if ctx.Scope.repo != "" {
		return ""
	}
	return ctx.Scope.last
}

// AddKnownRepos records repos seen in fetched rows.
func (ctx *ProgramContext) AddKnownRepos(repos ...string) {
	if !ctx.RepoScoped() {
		return
	}
	ctx.Scope.mu.Lock()
	defer ctx.Scope.mu.Unlock()
	for _, r := range repos {
		if r != "" {
			ctx.Scope.known[strings.ToLower(r)] = r
		}
	}
}

// KnownRepos returns the repos seen in fetched rows, sorted.
func (ctx *ProgramContext) KnownRepos() []string {
	if !ctx.RepoScoped() {
		return nil
	}
	ctx.Scope.mu.RLock()
	defer ctx.Scope.mu.RUnlock()
	repos := make([]string, 0, len(ctx.Scope.known))
	for _, r := range ctx.Scope.known {
		repos = append(repos, r)
	}
	slices.SortFunc(repos, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return repos
}

// RowRepoName returns how a row shows its repo: nothing when the scope is a
// single repo, the repo name when no other known repo has it, or the full
// name with the owner. It returns nameWithOwner when the scope is off.
func (ctx *ProgramContext) RowRepoName(nameWithOwner string) string {
	if !ctx.RepoScoped() {
		return nameWithOwner
	}
	if ctx.ScopeRepo() != "" {
		return ""
	}
	_, name, ok := strings.Cut(nameWithOwner, "/")
	if !ok {
		return nameWithOwner
	}
	ctx.Scope.mu.RLock()
	defer ctx.Scope.mu.RUnlock()
	for _, r := range ctx.Scope.known {
		if _, other, _ := strings.Cut(r, "/"); strings.EqualFold(other, name) &&
			!strings.EqualFold(r, nameWithOwner) {
			return nameWithOwner
		}
	}
	return name
}

// IsViewer tells if login is the signed-in user.
func (ctx *ProgramContext) IsViewer(login string) bool {
	return ctx != nil && ctx.User != "" && strings.EqualFold(ctx.User, login)
}

// WithScope adds the repo in scope to a search query, unless the query has
// its own repo qualifier.
func (ctx *ProgramContext) WithScope(query string) string {
	repo := ctx.ScopeRepo()
	if repo == "" {
		return query
	}
	for token := range strings.FieldsSeq(query) {
		if strings.HasPrefix(token, "repo:") {
			return query
		}
	}
	return strings.TrimSpace("repo:" + repo + " " + query)
}
