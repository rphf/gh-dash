package section

// MarkStale makes the section fetch again the next time its view is shown,
// and drops the results of the fetches in flight.
func (m *BaseModel) MarkStale() {
	m.LastFetch = FetchEvent{}
}
