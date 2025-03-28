package table

import (
	"fmt"
	"os"
	"slices"
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

	// Attributes of the table
	width    int
	widths   []int
	isEmpty  map[int]bool
	rowStyle map[int]*CellStyle
	colStyle map[int]*CellStyle
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
		isEmpty:  make(map[int]bool),
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

	headerWidths, minWidths, maxWidths := t.measureTable()

	t.autoResize(headerWidths, minWidths, maxWidths)

	// render header
	t.renderColumn(b, headerRow, t.header)

	// render rows
	for i, row := range t.rows {
		t.renderColumn(b, i, row)
	}

	return b.String()
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

func (t *table) autoResize(headerWidths, minWidths, maxWidths []int) {
	minSum := t.sumWidths(minWidths)
	maxSum := t.sumWidths(maxWidths)

	width := t.width - t.extraWidth()
	if width >= maxSum {
		t.widths = maxWidths
		return
	}

	t.widths = minWidths
	if width >= minSum {
		t.expandWidths(maxWidths, width-minSum)
	} else {
		t.shrinkWidths(headerWidths, minSum-width)
	}
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
			if t.style.HideEmpty && t.isEmpty[col] {
				continue
			}
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

func (t *table) measureTable() (headerWidths, minWidths, maxWidths []int) {
	headerWidths = make([]int, 0, len(t.header))
	minWidths = make([]int, 0, len(t.header))
	maxWidths = make([]int, 0, len(t.header))

	for col, h := range t.header {
		headerWidth := text.StringWidth(h)
		minWidth := headerWidth
		maxWidth := minWidth
		t.isEmpty[col] = true
		isWrap := t.style.WrapText

		for i, row := range t.rows {
			minCellWidth, maxCellWidth := t.measureCell(row[col])

			if minCellWidth != 0 {
				t.isEmpty[col] = false
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

		headerWidths = append(headerWidths, headerWidth)
		minWidths = append(minWidths, minWidth)
		maxWidths = append(maxWidths, maxWidth)
	}

	return
}

func (t *table) sumWidths(widths []int) int {
	var sum int
	for i, w := range widths {
		if t.style.HideEmpty && t.isEmpty[i] {
			continue
		}
		sum += w
	}
	return sum
}

func (t *table) extraWidth() int {
	cols := 0
	for i := range t.header {
		if t.style.HideEmpty && t.isEmpty[i] {
			continue
		}
		cols++
	}

	return t.style.OuterPadding*2 + t.style.InnerPadding*(cols-1)
}

type widthDiff struct {
	idx  int
	diff int
}

func (t *table) expandWidths(maxWidths []int, extra int) {
	ws := []*widthDiff{}
	for i, w := range t.widths {
		ws = append(ws, &widthDiff{idx: i, diff: maxWidths[i] - w})
	}

	slices.SortFunc(ws, func(a, b *widthDiff) int {
		return a.diff - b.diff
	})

	for _, w := range ws {
		if t.style.HideEmpty && t.isEmpty[w.idx] {
			continue
		}

		expended := min(w.diff, extra)
		t.widths[w.idx] += expended
		extra -= expended
		if extra == 0 {
			return
		}
	}
}

func (t *table) shrinkWidths(minWidths []int, extra int) {
	ws := []*widthDiff{}
	for i, w := range t.widths {
		ws = append(ws, &widthDiff{idx: i, diff: w - minWidths[i]})
	}

	slices.SortFunc(ws, func(a, b *widthDiff) int {
		return b.diff - a.diff
	})

	for _, w := range ws {
		if t.style.HideEmpty && t.isEmpty[w.idx] {
			continue
		}

		shrinked := min(w.diff, extra)
		t.widths[w.idx] -= shrinked
		extra -= shrinked
		if extra == 0 {
			return
		}
	}
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
