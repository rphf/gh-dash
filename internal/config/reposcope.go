package config

// RepoScopeConfig controls the repo scope: the repo whose PRs, issues and
// notifications are shown, or all repos.
type RepoScopeConfig struct {
	// Enabled shows one repo at a time, with a repo selector in the header,
	// and leaves the repo out of rows. It is true when not set.
	Enabled *bool `yaml:"enabled,omitempty"`
}

// IsEnabled tells if the repo scope is on.
func (c RepoScopeConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}
