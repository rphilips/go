package parse

import (
	mgold "github.com/yuin/goldmark"
	mast "github.com/yuin/goldmark/ast"
	mtext "github.com/yuin/goldmark/text"
)

func Parse(source []byte) mast.Node {
	reader := mtext.NewReader(source)
	md := mgold.New()
	parser := md.Parser()
	return parser.Parse(reader)
}

// reader := text.NewReader(source)
// doc := m.parser.Parse(reader, opts...)
// return m.renderer.Render(writer, source, doc)

// https://github.com/yuin/goldmark/blob/c71a97b8372876d63528b54cedecf1104530fe3b/markdown.go#L114
