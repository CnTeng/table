package table

import "strings"

type row []Cell

func (r row) render(ws widths) ([][]string, int) {
	cells := make([][]string, 0, len(r))

	maxRows := 0
	for colIdx, cell := range r {
		cellStr := cell.render(ws[colIdx])
		cells = append(cells, cellStr)
		if len(cellStr) > maxRows {
			maxRows = len(cellStr)
		}
	}

	for colIdx := range cells {
		if len(cells[colIdx]) < maxRows {
			for i := len(cells[colIdx]); i < maxRows; i++ {
				cells[colIdx] = append(cells[colIdx], strings.Repeat(" ", ws[colIdx]))
			}
		}
	}

	return cells, maxRows
}
