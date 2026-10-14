package issuessection

// onIssuesFetched runs after fetched issues are added to the section.
func (m *Model) onIssuesFetched() {
	m.ClampCursor(len(m.Issues))
	for _, issue := range m.Issues {
		m.Ctx.AddKnownRepos(issue.Repository.NameWithOwner)
	}
	if m.Ctx.RepoScoped() {
		m.Table.HideEmptyColumns(m.BuildRows())
	}
}
