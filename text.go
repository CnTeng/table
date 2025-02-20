package table

import (
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-runewidth"
)

// TODO: rewrite this
func wrapText(s string, width int) []string {
	return strings.Split(text.WrapSoft(s, width), "\n")
}

func rightJustify(s string, width int) string {
	inputWidth := runewidth.StringWidth(s)
	padding := width - inputWidth
	if padding > 0 {
		return strings.Repeat(" ", padding) + s
	}
	return s
}

func leftJustify(s string, width int) string {
	length := runewidth.StringWidth(stripansi.Strip(s))
	padding := width - length
	if padding > 0 {
		return s + strings.Repeat(" ", padding)
	}
	return s
}
