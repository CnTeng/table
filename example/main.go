package main

import (
	"fmt"
	"strings"

	"github.com/CnTeng/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func main() {
	tbl := table.NewTableWithStyle(&table.TableStyle{
		DefaultWidth:  80,
		FitToTerminal: false,
		WrapText:      true,
		Markdown:      true,
		HideEmpty:     true,
		OuterPadding:  0,
		InnerPadding:  1,
	})

	bgGreen := text.BgGreen.Sprint("BgGreen")
	bgBlue := text.BgBlue.Sprint("BgBlue")
	fgRed := text.FgRed.Sprint("FgRed")

	tbl.AddHeader("Feature", "Example", "Example")
	tbl.AddRow(table.Row{"**Color-256**", bgGreen + " " + bgBlue, fgRed})
	tbl.AddRow(table.Row{"**Wrap Text**", strings.Repeat("Hello World! ", 20), "Hello World!"})
	tbl.AddRow(table.Row{"Markdown", "[link](http://example.com)", "**Bold** _Italic_ `inline Code`"})

	tbl.SetHeaderStyle(&table.CellStyle{
		CellAttrs: text.Colors{text.FgGreen, text.Underline},
	})

	fmt.Print(tbl.Render())
}
