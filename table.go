// Package table offers an easy way to generate tables similar to Taskwarrior,
// featuring 256-color support, automatic text wrapping and Markdown rendering.
//
//	tbl := table.NewTableWithStyle(&table.TableStyle{
//		DefaultWidth:  80,
//		FitToTerminal: false,
//		WrapText:      true,
//		Markdown:      true,
//		HideEmpty:     true,
//		OuterPadding:  0,
//		InnerPadding:  1,
//	})
//
//	bgGreen := text.BgGreen.Sprint("BgGreen")
//	bgBlue := text.BgBlue.Sprint("BgBlue")
//	fgRed := text.FgRed.Sprint("FgRed")
//
//	tbl.AddHeader("Feature", "Example", "Example")
//	tbl.AddRow(table.Row{"**Color-256**", bgGreen + " " + bgBlue, fgRed})
//	tbl.AddRow(table.Row{"**Wrap Text**", strings.Repeat("Hello World! ", 20), "Hello World!"})
//	tbl.AddRow(table.Row{"Markdown", "[link](http://example.com)", "**Bold** _Italic_ `inline Code`"})
//
//	tbl.SetHeaderStyle(&table.CellStyle{
//		CellAttrs: text.Colors{text.FgGreen, text.Underline},
//	})
//
//	fmt.Print(tbl.Render())
package table

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

type Table interface {
	AddHeader(header ...string)
	AddRow(row Row)
	AddRows(rows []Row)

	Length() int

	SetStyle(style *TableStyle)
	SetHeaderStyle(style *CellStyle)
	SetRowStyle(row int, style *CellStyle)
	SetColStyle(col int, style *CellStyle)

	Render() string
}

const headerRow = -1

type Row []any

type table struct {
	// Style of the table
	style *TableStyle

	// Data of the table
	header row
	rows   []row

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
	width := style.DefaultWidth
	if style.FitToTerminal {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			width = w
		}
	}

	t.style = style
	t.width = width
}

func (t *table) AddHeader(header ...string) {
	for _, h := range header {
		t.header = append(t.header, Cell{Content: h})
	}
}

func (t *table) AddRow(r Row) {
	row := make(row, 0, len(r))
	for _, v := range r {
		switch v := v.(type) {
		case *Cell:
			row = append(row, *v)
		case Cell:
			row = append(row, v)
		case string:
			row = append(row, Cell{Content: v})
		default:
			row = append(row, Cell{Content: fmt.Sprint(v)})
		}
	}
	t.rows = append(t.rows, row)
}

func (t *table) AddRows(rows []Row) {
	for _, row := range rows {
		t.AddRow(row)
	}
}

func (t *table) Length() int {
	return len(t.rows)
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

	t.setCellStyle()
	emptyMap := t.measureTable()
	t.hideColumns(emptyMap)
	t.autoResize()

	// render header
	t.renderRow(b, t.header)

	// render rows
	for _, row := range t.rows {
		t.renderRow(b, row)
	}

	return b.String()
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

func (t *table) hideColumns(emptyMap map[int]bool) {
	if !t.style.HideEmpty {
		return
	}

	t.header = hideColumnsInRow(t.header, emptyMap)
	for i, row := range t.rows {
		t.rows[i] = hideColumnsInRow(row, emptyMap)
	}
	t.headerWidths = hideColumnsInRow(t.headerWidths, emptyMap)
	t.minWidths = hideColumnsInRow(t.minWidths, emptyMap)
	t.maxWidths = hideColumnsInRow(t.maxWidths, emptyMap)
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

func (t *table) setCellStyle() {
	for colIdx := range t.header {
		t.header[colIdx].style = t.cellStyle(headerRow, colIdx)
	}

	for rowIdx := range t.rows {
		for colIdx := range t.rows[rowIdx] {
			t.rows[rowIdx][colIdx].style = t.cellStyle(rowIdx, colIdx)
		}
	}
}

func (t *table) cellStyle(row, col int) *CellStyle {
	s := &CellStyle{
		WrapText: &t.style.WrapText,
		Markdown: &t.style.Markdown,
	}

	if row == headerRow {
		return s.merge(t.rowStyle[headerRow])
	}

	s.merge(t.rowStyle[row])
	s.merge(t.colStyle[col])
	s.merge(t.rows[row][col].style)
	return s
}

func (t *table) renderRow(b *strings.Builder, r row) {
	cells, lines := r.render(t.widths)

	for i := range lines {
		b.WriteString(strings.Repeat(" ", t.style.OuterPadding))
		for col, cell := range cells {
			b.WriteString(cell[i])
			if col < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", t.style.InnerPadding))
			}
		}
		b.WriteString(strings.Repeat(" ", t.style.OuterPadding))
		b.WriteByte('\n')
	}
}

func (t *table) measureTable() (emptyMap map[int]bool) {
	t.headerWidths = make(widths, 0, len(t.header))
	t.minWidths = make(widths, 0, len(t.header))
	t.maxWidths = make(widths, 0, len(t.header))
	emptyMap = make(map[int]bool, len(t.header))

	for col, h := range t.header {
		headerWidth := text.StringWidth(h.Content)
		minWidth := headerWidth
		maxWidth := minWidth
		emptyMap[col] = true
		isWrap := t.style.WrapText

		for i, row := range t.rows {
			minCellWidth, maxCellWidth := row[col].measure()

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
