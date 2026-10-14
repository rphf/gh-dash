package issuessection

// onIssuesFetched runs after fetched issues are added to the section.
func (m *Model) onIssuesFetched() {
	for _, issue := range m.Issues {
		m.Ctx.AddKnownRepos(issue.Repository.NameWithOwner)
	}
}
