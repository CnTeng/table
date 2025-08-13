package table

import (
	"strings"
	"testing"
)

func TestTableRender(t *testing.T) {
	emptyTbl := NewTable()

	normalTbl := NewTable()
	normalTbl.AddHeader("Header1", "Header2")
	normalTbl.AddRows([]Row{
		{"Row1Col1", "Row1Col2"},
		{"Row2Col1", "Row2Col2"},
	})

	wrappedTbl := NewTableWithStyle(&TableStyle{
		DefaultWidth:  32,
		FitToTerminal: false,
		WrapText:      true,
		InnerPadding:  1,
	})
	wrappedTbl.AddHeader("Header1", "Header2")
	wrappedTbl.AddRows([]Row{
		{"This is a long text that should wrap", "Row1"},
		{"Another long text that should wrap", "Row2"},
	})

	markdownTbl := NewTableWithStyle(&TableStyle{
		DefaultWidth:  32,
		FitToTerminal: false,
		Markdown:      true,
		InnerPadding:  1,
	})
	markdownTbl.AddHeader("Header1", "Header2")
	markdownTbl.AddRows([]Row{
		{"**Bold Text**", "~~Strikethrough~~"},
		{"[Link](https://example.com)", "Inline `code`"},
		{"*Italic Text*", "~~**Bold and Strikethrough**~~"},
	})

	tests := []struct {
		name string
		in   Table
		want string
	}{
		{
			name: "Empty Table",
			in:   emptyTbl,
			want: "",
		},
		{
			name: "Normal Table",
			in:   normalTbl,
			want: strings.Join([]string{
				"Header1  Header2 ",
				"Row1Col1 Row1Col2",
				"Row2Col1 Row2Col2\n",
			}, "\n"),
		},
		{
			name: "Wrapped Table",
			in:   wrappedTbl,
			want: strings.Join([]string{
				"Header1                  Header2",
				"This is a long text that Row1   ",
				"should wrap                     ",
				"Another long text that   Row2   ",
				"should wrap                     \n",
			}, "\n"),
		},
		{
			name: "Table with Markdown",
			in:   markdownTbl,
			want: strings.Join([]string{
				"Header1   Header2               ",
				"\x1b[1mBold Text\x1b[0m \x1b[9mStrikethrough\x1b[0m         ",
				"\x1b[1mLink\x1b[0m \x1b[4mhttps://example.com\x1b[0m Inline \x1b[1mcode\x1b[0m           ",
				"\x1b[3mItalic Text\x1b[0m \x1b[9m\x1b[1mBold and Strikethrough\x1b[0m\x1b[9m\x1b[0m\n",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Render(); got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}
