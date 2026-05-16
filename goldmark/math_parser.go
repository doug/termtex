package goldmark

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// mathInlineParser parses inline math: $...$
type mathInlineParser struct{}

func (p *mathInlineParser) Trigger() []byte {
	return []byte{'$'}
}

// parse implements the Pandoc TeX-math delimiter rules (also used by
// KaTeX's markdown auto-renderer):
//
//  1. The opening `$` must have a non-space character immediately to
//     its right.
//  2. The closing `$` must have a non-space character immediately to
//     its left.
//  3. The closing `$` must not be followed immediately by a digit.
//
// Together, rules 2 and 3 ensure that "$20,000 and $30,000" parses as
// prose rather than math. Rule 1 rejects "$ ... $" (typically a typo).
// A backslash-escaped dollar (`\$`) inside the expression doesn't
// close the math, matching Pandoc's behavior.
//
// Reference: https://pandoc.org/MANUAL.html#math
func (p *mathInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 2 || line[0] != '$' {
		return nil
	}
	// Skip $$ (display math handled by block parser).
	if line[1] == '$' {
		return nil
	}
	// Rule 1: opener can't be followed by whitespace.
	if isMathSpace(line[1]) {
		return nil
	}
	// Walk forward looking for the first `$` that satisfies the closer
	// rules. Skip `\$` so an escaped dollar inside the expression
	// doesn't close the math.
	for i := 2; i < len(line); i++ {
		if line[i] == '\\' && i+1 < len(line) {
			i++ // skip the escaped char
			continue
		}
		if line[i] != '$' {
			continue
		}
		// Rule 2: closer can't be preceded by whitespace.
		if isMathSpace(line[i-1]) {
			continue
		}
		// Rule 3: closer can't be followed by an ASCII digit.
		if i+1 < len(line) && line[i+1] >= '0' && line[i+1] <= '9' {
			continue
		}
		expr := line[1:i]
		if len(bytes.TrimSpace(expr)) == 0 {
			return nil
		}
		node := &MathInline{Expression: make([]byte, len(expr))}
		copy(node.Expression, expr)
		block.Advance(i + 1)
		return node
	}
	return nil
}

func isMathSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// texInlineParser recognizes KaTeX/Pandoc-style TeX delimiters:
// `\(...\)` for inline math and `\[...\]` for display math. These have
// no ambiguity with prose, so no spacing or digit rules are needed —
// just a literal closer search on the same line.
type texInlineParser struct{}

func (p *texInlineParser) Trigger() []byte {
	return []byte{'\\'}
}

func (p *texInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '\\' {
		return nil
	}
	var closer string
	switch line[1] {
	case '(':
		closer = `\)`
	case '[':
		closer = `\]`
	default:
		return nil
	}
	rest := line[2:]
	pos := bytes.Index(rest, []byte(closer))
	if pos < 0 {
		return nil
	}
	expr := rest[:pos]
	if len(bytes.TrimSpace(expr)) == 0 {
		return nil
	}
	node := &MathInline{Expression: make([]byte, len(expr))}
	copy(node.Expression, expr)
	block.Advance(2 + pos + 2)
	return node
}

// mathBlockParser parses display math blocks: $$...$$
type mathBlockParser struct{}

func (p *mathBlockParser) Trigger() []byte {
	return []byte{'$'}
}

func (p *mathBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)
	if !bytes.HasPrefix(trimmed, []byte("$$")) {
		return nil, parser.NoChildren
	}

	// Check if it's a single-line $$...$$. Mark the block closed so the
	// Continue() loop returns immediately instead of slurping subsequent
	// lines into the math expression.
	rest := trimmed[2:]
	if end := bytes.Index(rest, []byte("$$")); end >= 0 {
		expr := rest[:end]
		node := &MathBlock{Expression: bytes.TrimSpace(expr), closed: true}
		reader.AdvanceLine()
		return node, parser.NoChildren
	}

	// Multi-line: opening $$, content follows
	node := &MathBlock{}
	reader.AdvanceLine()
	return node, parser.NoChildren
}

func (p *mathBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	mb := node.(*MathBlock)
	if mb.closed {
		// Single-line $$...$$ already consumed by Open().
		return parser.Close
	}
	line, _ := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)
	if bytes.HasPrefix(trimmed, []byte("$$")) {
		reader.AdvanceLine()
		return parser.Close
	}

	if len(mb.Expression) > 0 {
		mb.Expression = append(mb.Expression, '\n')
	}
	mb.Expression = append(mb.Expression, bytes.TrimRight(line, "\n\r")...)
	reader.AdvanceLine()
	return parser.Continue | parser.NoChildren
}

func (p *mathBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	mb := node.(*MathBlock)
	mb.Expression = bytes.TrimSpace(mb.Expression)
}

func (p *mathBlockParser) CanInterruptParagraph() bool { return true }
func (p *mathBlockParser) CanAcceptIndentedLine() bool { return false }
