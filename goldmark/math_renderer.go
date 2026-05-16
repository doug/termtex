package goldmark

import (
	"strings"

	"github.com/doug/termtex"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// mathRenderer renders math AST nodes using termtex.
type mathRenderer struct {
	Style termtex.Style
}

func (r *mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.renderMathInline)
	reg.Register(KindMathBlock, r.renderMathBlock)
}

func (r *mathRenderer) renderMathInline(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*MathInline)
	rendered, err := termtex.Render(string(n.Expression), r.Style)
	if err != nil {
		// Fall back to raw expression
		w.WriteString("$")
		w.Write(n.Expression)
		w.WriteString("$")
		return ast.WalkContinue, nil
	}
	w.WriteString(rendered)
	return ast.WalkContinue, nil
}

func (r *mathRenderer) renderMathBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*MathBlock)
	rendered, err := termtex.Render(string(n.Expression), r.Style)
	if err != nil {
		w.WriteString("$$")
		w.Write(n.Expression)
		w.WriteString("$$")
		w.WriteByte('\n')
		return ast.WalkContinue, nil
	}

	// Display math: indent and add newlines for separation
	w.WriteByte('\n')
	for _, line := range strings.Split(rendered, "\n") {
		w.WriteString("  ")
		w.WriteString(line)
		w.WriteByte('\n')
	}
	w.WriteByte('\n')
	return ast.WalkContinue, nil
}
