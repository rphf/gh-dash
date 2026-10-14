package config

// AttentionConfig controls the signal that tells which PRs need the viewer.
type AttentionConfig struct {
	// Enabled shows, on each PR row, if it needs the viewer or is ready to
	// ship. It is true when not set.
	Enabled *bool `yaml:"enabled,omitempty"`
	// SortFirst puts the PRs that need the viewer first in each section, then
	// the PRs ready to ship. It is true when not set.
	SortFirst *bool `yaml:"sortFirst,omitempty"`
	// MyAuthors holds other logins whose PRs count as the viewer's, such as
	// the viewer's agents ("my-agent" or "app/my-agent").
	MyAuthors []string `yaml:"myAuthors,omitempty"`
}

// IsEnabled tells if the attention signal is on.
func (c AttentionConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

// IsSortFirst tells if PRs are sorted by attention.
func (c AttentionConfig) IsSortFirst() bool {
	return c.IsEnabled() && (c.SortFirst == nil || *c.SortFirst)
}
