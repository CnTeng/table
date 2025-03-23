package table

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

type Style struct {
	*color.Color
}

type Table interface {
	AddHeader(header ...string)
	AddRow(vals ...any)

	SetHeaderStyle(style ...color.Attribute)
	SetRowStyle(row int, style ...color.Attribute)
	SetColStyle(col int, style ...color.Attribute)

	Render() string
}

type table struct {
	// Data of the table
	header []string
	rows   [][]string

	// Style of the table
	extraPadding int
	intraPadding int

	headerStyle []color.Attribute
	rowStyle    map[int][]color.Attribute
	colStyle    map[int][]color.Attribute

	// Attributes of the table
	width  int
	widths []int
}

func NewTable(w int, autoExpand bool) Table {
	if autoExpand {
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			w = width
		}
	}

	return &table{
		extraPadding: 0,
		intraPadding: 1,
		rowStyle:     make(map[int][]color.Attribute),
		colStyle:     make(map[int][]color.Attribute),
		width:        w,
	}
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

func (t *table) SetHeaderStyle(style ...color.Attribute) {
	t.headerStyle = style
}

func (t *table) SetRowStyle(row int, style ...color.Attribute) {
	t.rowStyle[row] = style
}

func (t *table) SetColStyle(col int, style ...color.Attribute) {
	t.colStyle[col] = style
}

func (t *table) Render() string {
	var b strings.Builder

	minWidth, maxWidth := t.measureTable()

	// TODO: is print empty columns
	t.autoResize(minWidth, maxWidth)

	// render header
	t.renderColumn(&b, -1, t.header)

	// render rows
	for i, row := range t.rows {
		t.renderColumn(&b, i, row)
	}

	return b.String()
}

func (t *table) cellStyle(row, col int) Style {
	style := Style{color.New()}

	if row == -1 {
		style.Add(t.headerStyle...)
	}

	if s, ok := t.rowStyle[row]; ok {
		style.Add(s...)
	}

	if s, ok := t.colStyle[col]; ok {
		style.Add(s...)
	}

	return style
}

func (t *table) autoResize(minWidth, maxWidth []int) []int {
	minSum := sum(minWidth)
	maxSum := sum(maxWidth)

	widths := make([]int, 0, len(t.header))

	overage := t.width - t.intraPadding*(len(t.header)-1) - minSum
	if maxSum+t.intraPadding <= t.width {
		widths = maxWidth
	} else if overage < 0 {
		longest := 0
		secondLongest := 0
		for i, w := range minWidth {
			if w > minWidth[longest] {
				secondLongest = longest
				longest = i
			} else if w > minWidth[secondLongest] {
				secondLongest = i
			}
		}

		// Case 1: Shorten the longest column
		widths = minWidth
		if minWidth[longest] >= widths[secondLongest]-overage {
			minWidth[longest] += overage
		} else {
			dec := widths[longest] - widths[secondLongest]
			widths[longest] -= dec
			overage += dec

			half := overage/2 + overage%2
			minWidth[longest] += half
			minWidth[secondLongest] += half
		}
	} else if overage == 0 {
		widths = minWidth
	}
	t.widths = widths

	return widths
}

func (t *table) renderCell(s string, width int) []string {
	lines := strings.Split(text.WrapSoft(s, width), "\n")
	for i := range lines {
		// TODO: add more alignment
		lines[i] = text.AlignLeft.Apply(lines[i], width)
	}
	return lines
}

func (t *table) renderColumn(b *strings.Builder, row int, col []string) {
	cells := make([][]string, 0, len(col))
	maxLines := 0
	for i, c := range col {
		cell := t.renderCell(c, t.widths[i])
		if len(cell) > maxLines {
			maxLines = len(cell)
		}
		cells = append(cells, cell)
	}

	for i := range maxLines {
		for col, cell := range cells {
			if i < len(cell) {
				style := t.cellStyle(row, col)
				b.WriteString(style.Sprint(cell[i]))
			} else {
				b.WriteString(strings.Repeat(" ", t.widths[col]))
			}

			// TODO: more style
			if col < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", t.intraPadding))
			}
		}
		b.WriteByte('\n')
	}
}

func (t *table) measureCell(data string) (minWidth int, maxWidth int) {
	striped := text.StripEscape(data)

	minWidth = longestWord(striped)
	maxWidth = longestLine(striped)
	return
}

func (t *table) measureTable() (minWidths []int, maxWidths []int) {
	minWidths = make([]int, 0, len(t.header))
	maxWidths = make([]int, 0, len(t.header))

	for col, h := range t.header {
		minWidth := text.StringWidth(h)
		maxWidth := minWidth

		for _, row := range t.rows {
			mini, ideal := t.measureCell(row[col])

			if mini > minWidth {
				minWidth = mini
			}
			if ideal > maxWidth {
				maxWidth = ideal
			}
		}

		minWidths = append(minWidths, minWidth)
		maxWidths = append(maxWidths, maxWidth)
	}

	return
}

func longestLine(s string) int {
	longest := 0
	length := 0

	for _, r := range s {
		if r == '\n' {
			if length > longest {
				longest = length
			}
			length = 0
		} else {
			length += text.RuneWidth(r)
		}
	}

	if length > longest {
		longest = length
	}

	return longest
}

func longestWord(s string) int {
	longest := 0
	length := 0

	for _, r := range s {
		if unicode.IsSpace(r) {
			if length > longest {
				longest = length
			}
			length = 0
		} else {
			length += text.RuneWidth(r)
		}
	}

	if length > longest {
		longest = length
	}

	return longest
}

func sum(slice []int) int {
	var sum int
	for _, s := range slice {
		sum += s
	}
	return sum
}
