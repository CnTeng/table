package table

import (
	"bytes"
	"io"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.Strikethrough),
	goldmark.WithRenderer(newAnsiRenderer()),
)

func renderMarkdown(s string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		return s
	}
	return buf.String()
}

type ansiRenderer struct{}

func newAnsiRenderer() *ansiRenderer {
	return &ansiRenderer{}
}

func (r *ansiRenderer) AddOptions(...renderer.Option) {}

func (r *ansiRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	return ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		case *ast.Text:
			if entering {
				_, _ = w.Write(node.Segment.Value(source))
			}
		case *ast.Emphasis:
			if entering {
				switch node.Level {
				case 1:
					_, _ = w.Write([]byte(text.Italic.EscapeSeq()))
				case 2:
					_, _ = w.Write([]byte(text.Bold.EscapeSeq()))
				}
			} else {
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
			}
		case *east.Strikethrough:
			if entering {
				_, _ = w.Write([]byte(text.CrossedOut.EscapeSeq()))
			} else {
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
			}
		}
		return ast.WalkContinue, nil
	})
}
