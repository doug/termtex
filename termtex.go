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
// value is the package default (plain Unicode, no italic, no color,
// display style).
//
//	out, err := termtex.Render(input, termtex.Style{
//	    Italic: true,
//	    Color:  true,
//	})
//
// # Display and inline typesetting
//
// Layout follows TeX's math styles. The default is display style:
// fractions stack over a bar and the limits of \sum, \prod and \lim
// sit above and below the operator. Style.Inline selects text style,
// where fractions flatten to a/b (parenthesised only where needed)
// and limits become side scripts, so ordinary expressions stay on one
// row inside a sentence. Integrals take side limits in both styles,
// as in TeX; \limits and \nolimits override, and \dfrac / \tfrac
// force a stacked or flat fraction.
//
// Spacing between atoms comes from TeX's class table (Ord, Op, Bin,
// Rel, Open, Close, Punct, Inner): relations and binary operators get
// a cell on each side, punctuation a cell after, named functions a
// cell before an operand but none before a parenthesis, and a leading
// or post-relation minus is unary. Inside scripts the optional spaces
// are dropped.
//
// # Markdown integration
//
// [Expand] rewrites $...$ and $$...$$ in a markdown string to
// pre-rendered termtex output, inline math in text style and display
// math in display style. The result feeds cleanly into terminal
// markdown renderers like glamour. For custom goldmark pipelines, see
// the goldmark subpackage.
//
// # Supported LaTeX
//
// Fractions, binomials, super/subscripts, square and nth roots, big
// operators (\sum, \prod, \int and their families, \lim), Greek
// letters, math fonts (\mathbb, \mathcal, \mathbf, \mathfrak, \mathsf,
// \mathit), tall delimiters including \langle, \lfloor, \lceil and \|,
// matrix environments, equation arrays (align, aligned, gather, array,
// cases, and bare \\ / & at top level), accents (\hat, \tilde, \dot,
// \ddot, \vec) using combining marks, \overbrace / \underbrace, and the
// common operator and arrow set. See README.md for the full table.
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
	n = breakLines(n, ctx, style.Width)
	b := measure(n, ctx)
	c := newCanvas(b.Width, b.Height)
	c.ctx = ctx
	renderNode(c, n, 0, 0)
	out := c.String()
	c.release()
	return out, nil
}
