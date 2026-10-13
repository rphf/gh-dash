package config

// StacksConfig controls how stacked pull requests are shown in PR sections.
type StacksConfig struct {
	// Enabled groups stacked PRs together, in stack order, with a tree prefix.
	// It is true when not set.
	Enabled *bool `yaml:"enabled,omitempty"`
}

// IsEnabled tells if stacked PRs are shown as stacks.
func (c StacksConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}
