package prssection

// onPrsFetched runs after fetched PRs are added to the section. firstPage
// tells if they replaced the previous PRs.
func (m *Model) onPrsFetched(firstPage bool) {
	for _, pr := range m.Prs {
		m.Ctx.AddKnownRepos(pr.Primary.Repository.NameWithOwner)
	}
	m.arrangeStacks(firstPage)
	if m.Ctx.RepoScoped() {
		m.Table.HideEmptyColumns(m.BuildRows())
	}
}
