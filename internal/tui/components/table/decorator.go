package table

// RowDecorator changes a rendered row before it is shown, such as its
// separator line. It returns the row unchanged when it has nothing to add.
type RowDecorator func(rowId int, row string) string

// SetRowDecorator sets the decorator applied to each rendered row. A nil
// decorator leaves rows unchanged.
func (m *Model) SetRowDecorator(d RowDecorator) {
	m.rowDecorator = d
}

func (m *Model) decorateRow(rowId int, row string) string {
	if m.rowDecorator == nil {
		return row
	}
	return m.rowDecorator(rowId, row)
}
