package prssection

// onPrsFetched runs after fetched PRs are added to the section. firstPage
// tells if they replaced the previous PRs.
func (m *Model) onPrsFetched(firstPage bool) {
	for _, pr := range m.Prs {
		m.Ctx.AddKnownRepos(pr.Primary.Repository.NameWithOwner)
	}
	var selected string
	if cur := m.Table.GetCurrItem(); cur >= 0 && cur < len(m.rows.shown) {
		selected = m.rows.shown[cur]
	}
	m.arrangeStacks(firstPage)
	m.moveCursorTo(selected)
	if m.Ctx.RepoScoped() {
		m.Table.HideEmptyColumns(m.BuildRows())
	}
}

// moveCursorTo moves the cursor to the PR at url, or keeps it in the rows
// when the PR is not shown.
func (m *Model) moveCursorTo(url string) {
	target := -1
	for i, pr := range m.Prs {
		if pr.Primary.Url == url {
			target = i
		}
	}
	m.Table.SetRows(m.BuildRows())
	if target == -1 {
		m.ClampCursor(len(m.Prs))
		return
	}
	for cur := m.Table.GetCurrItem(); cur != target; {
		if cur < target {
			m.Table.NextItem()
		} else {
			m.Table.PrevItem()
		}
		if m.Table.GetCurrItem() == cur {
			return
		}
		cur = m.Table.GetCurrItem()
	}
}
