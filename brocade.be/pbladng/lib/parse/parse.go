package parse

import (
	pgold "github.com/yuin/goldmark"
	past "github.com/yuin/goldmark/ast"
	ptext "github.com/yuin/goldmark/text"
)

func Parse(source []byte) past.Node {
	reader := ptext.NewReader(source)
	md := pgold.New()
	parser := md.Parser()
	return parser.Parse(reader)
}

// reader := text.NewReader(source)
// doc := m.parser.Parse(reader, opts...)
// return m.renderer.Render(writer, source, doc)

// https://github.com/yuin/goldmark/blob/c71a97b8372876d63528b54cedecf1104530fe3b/markdown.go#L114
