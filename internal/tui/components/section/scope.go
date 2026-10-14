package section

// ClampCursor moves the cursor to the last of n rows when it is past them.
func (m *BaseModel) ClampCursor(n int) {
	for m.Table.GetCurrItem() >= n && m.Table.GetCurrItem() > 0 {
		m.Table.PrevItem()
	}
}

// MarkStale makes the section fetch again the next time its view is shown,
// and drops the results of the fetches in flight.
func (m *BaseModel) MarkStale() {
	m.LastFetch = FetchEvent{}
}
