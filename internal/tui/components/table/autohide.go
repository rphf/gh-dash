package table

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// autoHidden is the Hidden value of the columns hidden by HideEmptyColumns,
// which tells them apart from the columns hidden by config.
var autoHidden = true

// HideEmptyColumns hides the columns whose cells are blank in every row, and
// shows them again when a row has content. It leaves the columns hidden by
// config and the grow column as they are, and changes nothing when there are
// no rows.
func (m *Model) HideEmptyColumns(rows []Row) {
	if len(rows) == 0 {
		return
	}
	changed := false
	for i := range m.Columns {
		c := &m.Columns[i]
		isAuto := c.Hidden == &autoHidden
		if (c.Hidden != nil && *c.Hidden && !isAuto) || (c.Grow != nil && *c.Grow) {
			continue
		}
		switch blank := columnIsBlank(rows, i); {
		case blank && !isAuto:
			c.Hidden = &autoHidden
			changed = true
		case !blank && isAuto:
			c.Hidden = nil
			changed = true
		}
	}
	if changed {
		m.cacheColumnWidths()
	}
}

func columnIsBlank(rows []Row, i int) bool {
	for _, row := range rows {
		if i < len(row) && strings.TrimSpace(ansi.Strip(row[i])) != "" {
			return false
		}
	}
	return true
}
