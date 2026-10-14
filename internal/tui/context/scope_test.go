package context

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

func newScopedCtx(smart bool) *ProgramContext {
	ctx := &ProgramContext{
		GHRepo: &repository.Repository{Host: "github.com", Owner: "o", Name: "app"},
		Config: &config.Config{SmartFilteringAtLaunch: smart},
	}
	ctx.InitRepoScope()
	return ctx
}

func TestInitRepoScope(t *testing.T) {
	require.Equal(t, "o/app", newScopedCtx(true).ScopeRepo())

	ctx := newScopedCtx(false)
	require.True(t, ctx.RepoScoped())
	require.Equal(t, "", ctx.ScopeRepo())
	require.Equal(t, "o/app", ctx.ToggledScopeRepo())

	off := false
	ctx = &ProgramContext{Config: &config.Config{RepoScope: config.RepoScopeConfig{Enabled: &off}}}
	ctx.InitRepoScope()
	require.False(t, ctx.RepoScoped())
	require.Equal(t, "is:open", ctx.WithScope("is:open"))
}

func TestToggledScopeRepo(t *testing.T) {
	ctx := newScopedCtx(true)
	require.Equal(t, "", ctx.ToggledScopeRepo())
	ctx.SetScopeRepo("o/lib")
	ctx.SetScopeRepo("")
	require.Equal(t, "o/lib", ctx.ToggledScopeRepo())
}

func TestWithScope(t *testing.T) {
	ctx := newScopedCtx(true)
	require.Equal(t, "repo:o/app is:open", ctx.WithScope("is:open"))
	require.Equal(t, "repo:o/app", ctx.WithScope(""))
	require.Equal(t, "repo:x/y is:open", ctx.WithScope("repo:x/y is:open"))
	ctx.SetScopeRepo("")
	require.Equal(t, "is:open", ctx.WithScope("is:open"))
}

func TestRowRepoName(t *testing.T) {
	ctx := newScopedCtx(true)
	require.Equal(t, "", ctx.RowRepoName("o/app"))

	ctx.SetScopeRepo("")
	ctx.AddKnownRepos("o/app", "o/lib", "fork/lib")
	require.Equal(t, "app", ctx.RowRepoName("o/app"))
	require.Equal(t, "o/lib", ctx.RowRepoName("o/lib"))
	require.Equal(t, []string{"fork/lib", "o/app", "o/lib"}, ctx.KnownRepos())
}

func TestIsViewer(t *testing.T) {
	ctx := &ProgramContext{User: "Me"}
	require.True(t, ctx.IsViewer("me"))
	require.False(t, ctx.IsViewer("other"))
	require.False(t, (&ProgramContext{}).IsViewer(""))
}
