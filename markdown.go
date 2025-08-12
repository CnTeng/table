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
	b := &bytes.Buffer{}
	if err := md.Convert([]byte(s), b); err != nil {
		return s
	}
	return b.String()
}

type ansiRenderer struct {
	styleStack []string
}

func newAnsiRenderer() renderer.Renderer {
	return &ansiRenderer{
		styleStack: []string{},
	}
}

func (r *ansiRenderer) pushStyle(seq string) {
	r.styleStack = append(r.styleStack, seq)
}

func (r *ansiRenderer) popStyle() {
	if len(r.styleStack) > 0 {
		r.styleStack = r.styleStack[:len(r.styleStack)-1]
	}
}

func (r *ansiRenderer) styles() string {
	s := ""
	for _, seq := range r.styleStack {
		s += seq
	}
	return s
}

func (r *ansiRenderer) AddOptions(...renderer.Option) {}

func (r *ansiRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	return ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		case *ast.Text:
			if entering {
				_, _ = w.Write(node.Segment.Value(source))
			}

		case *ast.Link:
			if entering {
				_, _ = w.Write([]byte(text.Bold.EscapeSeq()))
			} else {
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
				_, _ = w.Write([]byte(" "))
				_, _ = w.Write([]byte(text.Underline.EscapeSeq()))
				_, _ = w.Write(node.Destination)
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
			}

		case *ast.Emphasis:
			if entering {
				seq := ""
				switch node.Level {
				case 1:
					seq = text.Italic.EscapeSeq()
				case 2:
					seq = text.Bold.EscapeSeq()
				}
				r.pushStyle(seq)
				_, _ = w.Write([]byte(seq))
			} else {
				r.popStyle()
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
				_, _ = w.Write([]byte(r.styles()))
			}

		case *east.Strikethrough:
			if entering {
				seq := text.CrossedOut.EscapeSeq()
				r.pushStyle(seq)
				_, _ = w.Write([]byte(seq))
			} else {
				r.popStyle()
				_, _ = w.Write([]byte(text.Reset.EscapeSeq()))
				_, _ = w.Write([]byte(r.styles()))
			}
		}
		return ast.WalkContinue, nil
	})
}
