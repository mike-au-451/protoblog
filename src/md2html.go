package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func main() {
	md, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
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

	var out bytes.Buffer

	err = gm.Convert(md, &out)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("%s\n", out.String())
}