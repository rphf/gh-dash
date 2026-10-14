package keys

import "charm.land/bubbles/v2/key"

var repoScopeBinding = key.NewBinding(
	key.WithKeys("ctrl+r"),
	key.WithHelp("ctrl+r", "switch repo"),
)
