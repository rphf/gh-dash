package keys

import "charm.land/bubbles/v2/key"

var toggleStackBinding = key.NewBinding(
	key.WithKeys("z"),
	key.WithHelp("z", "fold/unfold stack"),
)
