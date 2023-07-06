/*
Wrap the goldmark markdown renderer
*/
package render

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark-highlighting/v2"

	// log "main/logger"
)

type Markdown struct {
	md goldmark.Markdown
}

func New() *Markdown {
	// log.Trace("render.New")
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
			highlighting.Highlighting,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &Markdown{ md }
}

// The signature is not the same as goldmark.Convert
func (m Markdown) Convert(source []byte, w io.Writer) error {
	// log.Trace("render.Convert")
	return m.md.Convert(source, w)
}

