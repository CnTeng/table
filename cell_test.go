package table

import (
	"reflect"
	"testing"

	"github.com/jedib0t/go-pretty/v6/text"
)

func boolPtr(b bool) *bool { return &b }

func TestCellMeasure_Markdown(t *testing.T) {
	tests := []struct {
		name    string
		in      *Cell
		wantMin int
		wantMax int
	}{
		{
			name: "Plain Text",
			in: &Cell{
				Content: "Hello World",
				style:   &CellStyle{},
			},
			wantMin: 5,
			wantMax: 11,
		},
		{
			name: "Prefix Suffix",
			in: &Cell{
				Content: "Hello World",
				Prefix:  "[",
				Suffix:  "]",
				style:   &CellStyle{},
			},
			wantMin: 5,
			wantMax: 13,
		},
		{
			name: "Markdown",
			in: &Cell{
				Content: "**Hello World**",
				style: &CellStyle{
					Markdown: boolPtr(true),
				},
			},
			wantMin: 5,
			wantMax: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if min, max := tt.in.measure(); min != tt.wantMin || max != tt.wantMax {
				t.Errorf("measure() = min: %d, max: %d; want min: %d, max: %d", min, max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCellRender(t *testing.T) {
	tests := []struct {
		name  string
		in    *Cell
		width int
		want  []string
	}{
		{
			name: "Simple Cell",
			in: &Cell{
				Content: "Hello",
				style:   &CellStyle{},
			},
			width: 5,
			want:  []string{"Hello"},
		},
		{
			name: "Prefix Suffix",
			in: &Cell{
				Content: "World",
				Prefix:  "[",
				Suffix:  "]",
				style:   &CellStyle{},
			},
			width: 7,
			want:  []string{"[World]"},
		},
		{
			name: "Markdown",
			in: &Cell{
				Content: "**Bold**",
				style:   &CellStyle{Markdown: boolPtr(true)},
			},
			width: 6,
			want:  []string{"\x1b[1mBold\x1b[0m  "},
		},
		{
			name: "Wrap Text",
			in: &Cell{
				Content: "This is a long text that should wrap",
				style:   &CellStyle{WrapText: boolPtr(true)},
			},
			width: 20,
			want: []string{
				"This is a long text ",
				"that should wrap    ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.measure()
			if got := tt.in.render(tt.width); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("render() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCellStyleMerge(t *testing.T) {
	a := &CellStyle{
		Align:     text.AlignLeft,
		WrapText:  nil,
		Markdown:  boolPtr(false),
		TextAttrs: text.Colors{text.Bold},
		CellAttrs: text.Colors{text.FgBlue},
	}
	b := &CellStyle{
		Align:     text.AlignRight,
		WrapText:  boolPtr(true),
		Markdown:  boolPtr(true),
		TextAttrs: text.Colors{text.Italic},
		CellAttrs: text.Colors{text.BgYellow},
	}
	merged := a.merge(b)

	want := &CellStyle{
		Align:     text.AlignRight,
		WrapText:  boolPtr(true),
		Markdown:  boolPtr(true),
		TextAttrs: text.Colors{text.Bold, text.Italic},
		CellAttrs: text.Colors{text.FgBlue, text.BgYellow},
	}
	if !reflect.DeepEqual(merged, want) {
		t.Errorf("merge() = %+v; want %+v", merged, want)
	}
}

func TestLongestLine(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"hello\nworld", 5},
		{"short\nmuchlongerline\nmid", 14},
		{"singleline", 10},
		{"", 0},
		{"a\nbb\nccc", 3},
		{"你好\n世界", 2 * text.RuneWidth('你')},
		{"a\nb\nc", 1},
	}

	for _, tt := range tests {
		if l := longestLine(tt.in); l != tt.want {
			t.Errorf("longestLine(%q) = %d; want %d", tt.in, l, tt.want)
		}
	}
}

func TestLongestWord(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"hello world", 5},
		{"short muchlonger mid", 10},
		{"singleword", 10},
		{"", 0},
		{"a bb ccc", 3},
		{"你好 世界", 2 * text.RuneWidth('你')},
		{"a b c", 1},
	}

	for _, tt := range tests {
		if l := longestWord(tt.in); l != tt.want {
			t.Errorf("longestWord(%q) = %d; want %d", tt.in, l, tt.want)
		}
	}
}
