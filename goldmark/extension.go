package goldmark

import (
	"github.com/doug/termtex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// MathExtension is a goldmark extension that adds LaTeX math support.
// It recognizes $...$ for inline math and $$...$$ for display math,
// rendering them with termtex.
type MathExtension struct {
	Style termtex.Style
}

// NewMathExtension creates a new math extension with the default
// (zero-value) style.
func NewMathExtension() *MathExtension {
	return &MathExtension{}
}

// NewMathExtensionWithStyle creates a new math extension with custom style.
func NewMathExtensionWithStyle(style termtex.Style) *MathExtension {
	return &MathExtension{Style: style}
}

// Extend implements goldmark.Extender.
func (e *MathExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&mathBlockParser{}, 500),
		),
		parser.WithInlineParsers(
			util.Prioritized(&mathInlineParser{}, 500),
			// Priority 100 so we beat the default backslash-escape
			// parser (which would otherwise eat `\(` as a literal `(`).
			util.Prioritized(&texInlineParser{}, 100),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&mathRenderer{Style: e.Style}, 500),
		),
	)
}
