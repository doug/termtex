// Package termtex renders LaTeX math expressions as Unicode text suitable
// for terminal display.
//
// It parses a subset of LaTeX math syntax and typesets it on a character
// grid using Unicode box-drawing characters, mathematical symbols, and
// (optionally) italic letter forms or ANSI color.
//
// # Quick start
//
//	out, err := termtex.Render(`\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`, termtex.Style{})
//	fmt.Println(out)
//
// # Style
//
// [Style] toggles italic letters, ANSI color, and a strict 7-bit ASCII
// fallback for environments without full Unicode support. Its zero
// value is the package default (plain Unicode, no italic, no color).
//
//	out, err := termtex.Render(input, termtex.Style{
//	    Italic: true,
//	    Color:  true,
//	})
//
// # Markdown integration
//
// [Expand] rewrites $...$ and $$...$$ in a markdown string to
// pre-rendered termtex output. The result feeds cleanly into terminal
// markdown renderers like glamour. For custom goldmark pipelines, see
// the goldmark subpackage.
//
// # Supported LaTeX
//
// Fractions, super/subscripts, square and nth roots, big operators
// (\sum, \prod, \int, \oint, \lim), Greek letters, math fonts (\mathbb,
// \mathcal, \mathbf, \mathfrak, \mathsf, \mathit), tall delimiters,
// matrix environments, accents (\hat, \tilde, \dot, \ddot, \vec) using
// combining marks, \overbrace / \underbrace, and the common operator
// and arrow set. See README.md for the full table.
package termtex

// Render parses a LaTeX math string and returns a multi-line Unicode
// string suitable for terminal display. Pass [Style]{} for the package
// default.
//
// Returns an error if the input is malformed.
func Render(input string, style Style) (string, error) {
	n, err := parse(input)
	if err != nil {
		return "", err
	}
	ctx := newRenderCtx(style)
	if hasMixedSimpleScripts(n, ctx) {
		ctx.forceStackScripts = true
	}
	b := measure(n, ctx)
	c := newCanvas(b.Width, b.Height)
	c.ctx = ctx
	renderNode(c, n, 0, 0)
	return c.String(), nil
}
