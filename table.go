package table

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

type Table interface {
	AddHeader(header ...string)
	AddRow(vals ...any)
	AddRows(rows ...[]any)

	SetStyle(style *TableStyle)
	SetHeaderStyle(style *CellStyle)
	SetRowStyle(row int, style *CellStyle)
	SetColStyle(col int, style *CellStyle)

	Render() string
}

const headerRow = -1

type table struct {
	// Style of the table
	style *TableStyle

	// Data of the table
	header []string
	rows   [][]string

	// Row and column styles
	rowStyle map[int]*CellStyle
	colStyle map[int]*CellStyle

	// Attributes of the table
	width        int
	widths       widths
	headerWidths widths
	minWidths    widths
	maxWidths    widths
}

func NewTable() Table {
	return NewTableWithStyle(defaultTableStyle)
}

func NewTableWithStyle(style *TableStyle) Table {
	width := style.DefaultWidth
	if style.FitToTerminal {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			width = w
		}
	}

	return &table{
		style:    style,
		width:    width,
		rowStyle: make(map[int]*CellStyle),
		colStyle: make(map[int]*CellStyle),
	}
}

func (t *table) SetStyle(style *TableStyle) {
	t.style = style
}

func (t *table) AddHeader(header ...string) {
	t.header = header
}

func (t *table) AddRow(vals ...any) {
	row := make([]string, 0, len(vals))
	for _, v := range vals {
		switch v := v.(type) {
		case string:
			row = append(row, v)
		default:
			row = append(row, fmt.Sprint(v))
		}
	}
	t.rows = append(t.rows, row)
}

func (t *table) AddRows(rows ...[]any) {
	for _, row := range rows {
		t.AddRow(row...)
	}
}

func (t *table) SetHeaderStyle(style *CellStyle) {
	t.rowStyle[headerRow] = style
}

func (t *table) SetRowStyle(row int, style *CellStyle) {
	t.rowStyle[row] = style
}

func (t *table) SetColStyle(col int, style *CellStyle) {
	t.colStyle[col] = style
}

func (t *table) Render() string {
	b := &strings.Builder{}

	emptyMap := t.measureTable()
	t.hideColumns(emptyMap)
	t.autoResize()

	// render header
	t.renderColumn(b, headerRow, t.header)

	// render rows
	for i, row := range t.rows {
		t.renderColumn(b, i, row)
	}

	return b.String()
}

func (t *table) hideColumns(emptyMap map[int]bool) {
	if !t.style.HideEmpty {
		return
	}
	colIdxMap := t.dropEmptyColumns(emptyMap)
	t.remapColStyle(colIdxMap)
}

func hideColumnsInRow[T any](row []T, emptyMap map[int]bool) []T {
	newRow := make([]T, 0, len(row))
	for colIdx, cell := range row {
		if emptyMap[colIdx] {
			continue
		}
		newRow = append(newRow, cell)
	}
	return newRow
}

func (t *table) dropEmptyColumns(emptyMap map[int]bool) map[int]int {
	colIdxMap := make(map[int]int)

	hideColumnsInHeader := func(row []string) []string {
		newRow := []string{}
		for colIdx, cell := range row {
			if emptyMap[colIdx] {
				continue
			}
			newRow = append(newRow, cell)
			colIdxMap[colIdx] = len(newRow) - 1
		}
		return newRow
	}

	t.header = hideColumnsInHeader(t.header)
	for i, row := range t.rows {
		t.rows[i] = hideColumnsInRow(row, emptyMap)
	}

	t.headerWidths = hideColumnsInRow(t.headerWidths, emptyMap)
	t.minWidths = hideColumnsInRow(t.minWidths, emptyMap)
	t.maxWidths = hideColumnsInRow(t.maxWidths, emptyMap)

	return colIdxMap
}

func (t *table) remapColStyle(colIdxMap map[int]int) {
	newColStyle := make(map[int]*CellStyle)
	for colIdx, style := range t.colStyle {
		if newColIdx, ok := colIdxMap[colIdx]; ok {
			newColStyle[newColIdx] = style
		}
	}
	t.colStyle = newColStyle
}

func (t *table) autoResize() {
	minSum := t.minWidths.sum()
	maxSum := t.maxWidths.sum()

	width := t.width - t.style.OuterPadding*2 - t.style.InnerPadding*(len(t.header)-1)
	if width >= maxSum {
		t.widths = t.maxWidths
		return
	}

	t.widths = t.minWidths
	if width >= minSum {
		t.widths.expand(t.maxWidths, width-minSum)
	} else {
		t.widths.shrink(t.headerWidths, minSum-width)
	}
}

func (t *table) cellStyle(row, col int) *CellStyle {
	s := &CellStyle{WrapText: &t.style.WrapText}

	if row == headerRow {
		return s.merge(t.rowStyle[headerRow])
	}

	s.merge(t.rowStyle[row])
	s.merge(t.colStyle[col])
	return s
}

func (t *table) renderColumn(b *strings.Builder, row int, cols []string) {
	cells := make([][]string, 0, len(cols))
	maxLines := 0
	for col, cell := range cols {
		cell := t.cellStyle(row, col).render(cell, t.widths[col])
		if len(cell) > maxLines {
			maxLines = len(cell)
		}
		cells = append(cells, cell)
	}

	for i := range maxLines {
		b.WriteString(strings.Repeat(" ", t.style.OuterPadding))
		for col, cell := range cells {
			if i < len(cell) {
				b.WriteString(cell[i])
			} else {
				b.WriteString(strings.Repeat(" ", t.widths[col]))
			}

			if col < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", t.style.InnerPadding))
			}
		}
		b.WriteString(strings.Repeat(" ", t.style.OuterPadding))
		b.WriteByte('\n')
	}
}

func (t *table) measureCell(data string) (minWidth int, maxWidth int) {
	striped := text.StripEscape(data)

	minWidth = longestWord(striped)
	maxWidth = longestLine(striped)
	return
}

func (t *table) measureTable() (emptyMap map[int]bool) {
	t.headerWidths = make(widths, 0, len(t.header))
	t.minWidths = make(widths, 0, len(t.header))
	t.maxWidths = make(widths, 0, len(t.header))
	emptyMap = make(map[int]bool, len(t.header))

	for col, h := range t.header {
		headerWidth := text.StringWidth(h)
		minWidth := headerWidth
		maxWidth := minWidth
		emptyMap[col] = true
		isWrap := t.style.WrapText

		for i, row := range t.rows {
			minCellWidth, maxCellWidth := t.measureCell(row[col])

			if minCellWidth != 0 {
				emptyMap[col] = false
			}

			s := t.cellStyle(i, col)
			if !*s.WrapText {
				isWrap = false
			}

			if minCellWidth > minWidth {
				minWidth = minCellWidth
			}
			if maxCellWidth > maxWidth {
				maxWidth = maxCellWidth
			}
		}

		if !isWrap {
			minWidth = maxWidth
		}

		t.headerWidths = append(t.headerWidths, headerWidth)
		t.minWidths = append(t.minWidths, minWidth)
		t.maxWidths = append(t.maxWidths, maxWidth)
	}

	return
}

func longestLine(s string) int {
	maxLength := 0
	curLength := 0

	for _, r := range s {
		if r == '\n' {
			if curLength > maxLength {
				maxLength = curLength
			}
			curLength = 0
		} else {
			curLength += text.RuneWidth(r)
		}
	}

	if curLength > maxLength {
		maxLength = curLength
	}

	return maxLength
}

func longestWord(s string) int {
	maxLength := 0
	curLength := 0

	for _, r := range s {
		if unicode.IsSpace(r) {
			if curLength > maxLength {
				maxLength = curLength
			}
			curLength = 0
		} else {
			curLength += text.RuneWidth(r)
		}
	}

	if curLength > maxLength {
		maxLength = curLength
	}

	return maxLength
}
