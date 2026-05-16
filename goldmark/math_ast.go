// Package goldmark provides a goldmark extension for rendering LaTeX math
// in the terminal using termtex.
//
// It recognizes $...$ for inline math and $$...$$ for display math blocks,
// parsing them into custom AST nodes that are rendered by calling termtex.
//
// Usage with goldmark directly:
//
//	md := goldmark.New(goldmark.WithExtensions(goldmark.NewMathExtension()))
//	md.Convert(source, os.Stdout)
//
// Usage with glamour (requires building a custom renderer — see example/):
//
//	// See the example program for the full glamour integration pattern.
package goldmark

import "github.com/yuin/goldmark/ast"

// KindMathInline is the AST node kind for inline math ($...$).
var KindMathInline = ast.NewNodeKind("MathInline")

// KindMathBlock is the AST node kind for display math ($$...$$).
var KindMathBlock = ast.NewNodeKind("MathBlock")

// MathInline represents an inline math expression.
type MathInline struct {
	ast.BaseInline
	Expression []byte
}

func (n *MathInline) Kind() ast.NodeKind { return KindMathInline }
func (n *MathInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// MathBlock represents a display math block.
type MathBlock struct {
	ast.BaseBlock
	Expression []byte
	closed     bool // set when the block parser already has the full expression
}

func (n *MathBlock) Kind() ast.NodeKind { return KindMathBlock }
func (n *MathBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
