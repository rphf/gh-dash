package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

// FetchTopRepos returns the repos the viewer contributes to, most recently
// updated first.
func FetchTopRepos(limit int) ([]string, error) {
	c := client
	if c == nil {
		var err error
		if c, err = gh.DefaultGraphQLClient(); err != nil {
			return nil, err
		}
	}

	var query struct {
		Viewer struct {
			TopRepositories struct {
				Nodes []struct {
					NameWithOwner string
				}
			} `graphql:"topRepositories(first: $limit, orderBy: {field: UPDATED_AT, direction: DESC})"`
		}
	}
	err := c.Query("TopRepositories", &query, map[string]any{"limit": graphql.Int(limit)})
	if err != nil {
		return nil, err
	}
	repos := make([]string, 0, len(query.Viewer.TopRepositories.Nodes))
	for _, n := range query.Viewer.TopRepositories.Nodes {
		repos = append(repos, n.NameWithOwner)
	}
	return repos, nil
}

const (
	recentReposFile = "recent-repos.json"
	maxRecentRepos  = 20
)

var recentReposMu sync.Mutex

// RecentRepos returns the repos last put in scope, most recent first.
func RecentRepos() []string {
	recentReposMu.Lock()
	defer recentReposMu.Unlock()
	return loadRecentRepos()
}

// AddRecentRepo puts repo first in the recent repos.
func AddRecentRepo(repo string) error {
	if repo == "" {
		return nil
	}
	recentReposMu.Lock()
	defer recentReposMu.Unlock()

	repos := []string{repo}
	for _, r := range loadRecentRepos() {
		if !strings.EqualFold(r, repo) && len(repos) < maxRecentRepos {
			repos = append(repos, r)
		}
	}

	path, err := getStateFilePath(recentReposFile)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func loadRecentRepos() []string {
	path, err := getStateFilePath(recentReposFile)
	if err != nil {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var repos []string
	if json.Unmarshal(b, &repos) != nil {
		return nil
	}
	return repos
}
