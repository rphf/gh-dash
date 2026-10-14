package keys

import "charm.land/bubbles/v2/key"

var (
	hidePRBinding = key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "hide/unhide PR"),
	)
	showHiddenPRsBinding = key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "show/hide hidden PRs"),
	)
)
