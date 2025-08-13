package table

import (
	"reflect"
	"testing"
)

func TestRowRender(t *testing.T) {
	r := row{
		Cell{
			Content: "Hello\nWorld",
			style:   &CellStyle{},
		},
		Cell{
			Content: "Hello World",
			style:   &CellStyle{},
		},
	}
	ws := widths{5, 12}

	gotCells, gotMaxRows := r.render(ws)

	wantCells := [][]string{
		{"Hello", "World"},
		{"Hello World ", "            "},
	}
	wantMaxRows := 2

	if !reflect.DeepEqual(gotCells, wantCells) {
		t.Errorf("render() = cells: %q, want cells: %q", gotCells, wantCells)
	}
	if gotMaxRows != wantMaxRows {
		t.Errorf("render() = maxRows: %q, want maxRows: %q", gotMaxRows, wantMaxRows)
	}
}
