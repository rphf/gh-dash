package notificationssection

// onNotificationsFetched runs after fetched notifications are added to the
// section.
func (m *Model) onNotificationsFetched() {
	for _, n := range m.Notifications {
		m.Ctx.AddKnownRepos(n.GetRepoNameWithOwner())
	}
}
