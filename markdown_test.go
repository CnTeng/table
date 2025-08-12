package table

import (
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "Bold",
			in:   "**bold**",
			want: "\x1b[1mbold\x1b[0m",
		},
		{
			name: "Italic",
			in:   "*italic*",
			want: "\x1b[3mitalic\x1b[0m",
		},
		{
			name: "Strikethrough",
			in:   "~~strike~~",
			want: "\x1b[9mstrike\x1b[0m",
		},
		{
			name: "Code",
			in:   "`inline code`",
			want: "\x1b[1minline code\x1b[0m",
		},
		{
			name: "Link",
			in:   "[link](http://example.com)",
			want: "\x1b[1mlink\x1b[0m \x1b[4mhttp://example.com\x1b[0m",
		},
		{
			name: "Nested Emphasis",
			in:   "This is ~~italic and **bold** text~~",
			want: "This is \x1b[9mitalic and \x1b[1mbold\x1b[0m\x1b[9m text\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if out := renderMarkdown(tt.in); out != tt.want {
				t.Errorf("renderMarkdown(%q) = %q, want %q", tt.in, out, tt.want)
			}
		})
	}
}
