package table

import (
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/text"
)

type Cell struct {
	Content string

	// Prefix defines the prefix of the cell.
	Prefix string
	// PrefixFunc defines the prefix function of the cell.
	PrefixFunc func(isFirst, isLast bool) string

	// Suffix defines the suffix of the cell.
	Suffix string
	// SuffixFunc defines the suffix function of the cell.
	SuffixFunc func(isFirst, isLast bool) string

	style *CellStyle
}

func (c *Cell) measure() (minWidth, maxWidth int) {
	c.Content = renderMarkdown(c.Content)
	striped := text.StripEscape(c.Content)

	minWidth = longestWord(striped)
	maxWidth = c.prefixLength() + longestLine(striped) + c.suffixLength()
	return
}

func (c *Cell) prefixLength() int {
	if c.Prefix != "" {
		return text.StringWidthWithoutEscSequences(c.Prefix)
	}
	if c.PrefixFunc != nil {
		return text.StringWidthWithoutEscSequences(c.PrefixFunc(true, false))
	}
	return 0
}

func (c *Cell) suffixLength() int {
	if c.Suffix != "" {
		return text.StringWidthWithoutEscSequences(c.Suffix)
	}
	if c.SuffixFunc != nil {
		return text.StringWidthWithoutEscSequences(c.SuffixFunc(true, false))
	}
	return 0
}

func (c *Cell) render(width int) []string {
	prefixLength := c.prefixLength()
	suffixLength := c.suffixLength()

	if prefixLength+suffixLength < width {
		width -= prefixLength + suffixLength
	}

	if *c.style.WrapText {
		c.Content = text.WrapSoft(c.Content, width)
	}

	lines := strings.Split(c.Content, "\n")
	for i, line := range lines {
		line = c.style.TextAttrs.Sprint(line)
		line = c.style.Align.Apply(line, width)
		if c.Prefix != "" {
			line = c.Prefix + line
		} else if c.PrefixFunc != nil {
			line = c.PrefixFunc(i == 0, i == len(lines)-1) + line
		}
		if c.Suffix != "" {
			line = line + c.Suffix
		} else if c.SuffixFunc != nil {
			line = line + c.SuffixFunc(i == 0, i == len(lines)-1)
		}
		line = c.style.CellAttrs.Sprint(line)

		lines[i] = line
	}

	return lines
}

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
