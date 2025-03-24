package table

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

type TableStyle struct {
	// DefaultWidth defines the default width of the table.
	DefaultWidth int
	// FitToTerminal defines if the table should fit to the terminal width.
	FitToTerminal bool

	// WrapText defines if the text should be wrapped.
	WrapText bool
	// HideEmpty defines if empty rows should be hidden.
	HideEmpty bool

	// OuterPadding defines the padding around the table.
	OuterPadding int
	// InnerPadding defines the padding between the cells.
	InnerPadding int
}

var defaultTableStyle = &TableStyle{
	DefaultWidth:  80,
	FitToTerminal: true,
	WrapText:      true,
	HideEmpty:     true,
	OuterPadding:  0,
	InnerPadding:  1,
}

func BoolPtr(b bool) *bool { return &b }

// CellStyle is the style of a cell in the table
type CellStyle struct {
	// Align defines the alignment of the text.
	Align text.Align

	// WrapText defines if the text should be wrapped.
	WrapText *bool

	// TextAttrs defines the text attributes.
	TextAttrs text.Colors

	// CellAttrs defines the cell attributes.
	CellAttrs text.Colors
}

func (cs *CellStyle) merge(other *CellStyle) *CellStyle {
	if other == nil {
		return cs
	}

	if cs == nil {
		return other
	}

	cs.Align = other.Align
	if other.WrapText != nil {
		cs.WrapText = other.WrapText
	}
	cs.TextAttrs = append(cs.TextAttrs, other.TextAttrs...)
	cs.CellAttrs = append(cs.CellAttrs, other.CellAttrs...)
	removeDuplicates(cs.TextAttrs)
	removeDuplicates(cs.CellAttrs)

	return cs
}

func (cs *CellStyle) render(s string, width int) []string {
	if *cs.WrapText {
		s = text.WrapSoft(s, width)
	}

	lines := make([]string, 0)
	for line := range strings.SplitSeq(s, "\n") {
		line = cs.TextAttrs.Sprint(line)
		line = cs.Align.Apply(line, width)
		line = cs.CellAttrs.Sprint(line)

		lines = append(lines, line)
	}

	return lines
}

func removeDuplicates[S ~[]E, E comparable](s S) S {
	if len(s) == 0 {
		return s
	}

	seen := make(map[E]any)
	result := S{}

	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
