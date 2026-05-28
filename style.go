package termtex

import "fmt"

// Style controls rendering options. The zero value is the package
// default: plain Unicode, no italic, no color.
type Style struct {
	// Italic uses the Mathematical Italic Unicode block (U+1D434…) for
	// variable letters. Requires a font with coverage of that range —
	// most stock monospace fonts do not include it.
	Italic bool
	// Color emits ANSI 24-bit color escapes for variables, numbers,
	// operators, delimiters, and big operators.
	Color bool
	// ASCII restricts output to 7-bit ASCII, falling back to plain
	// characters for box drawing, fraction bars, sqrt, accents, etc.
	// Useful for code comments, CI logs, or terminals lacking Unicode.
	ASCII bool
}

// renderCtx bundles user-facing [Style] options with derived runtime
// state used during a single measure-and-render pass. Keeping these
// separate from [Style] avoids exposing pass-internal flags on the
// public API.
type renderCtx struct {
	Style
	// forceStackScripts disables inline Unicode super/subscripts so all
	// scripts in a single render pass stack uniformly. Set by
	// renderTree when any script in the AST cannot be inlined,
	// to avoid mixing inline (`Tₕ`) and stacked (`T \n c`) forms in the
	// same expression.
	forceStackScripts bool
	// compact is set when rendering script content (sub/sup, big-op
	// limits) where operator spacing should be suppressed (e.g. `i=1`
	// instead of `i = 1`). Replaces the older structural heuristic in
	// groupNeedsSpacing.
	compact bool
}

// withCompact returns a copy of the context with compact=true.
func (c renderCtx) withCompact() renderCtx {
	c.compact = true
	return c
}

// newRenderCtx wraps a [Style] in a fresh render context with no
// derived state set. Measurement memoization lives on the AST node.
func newRenderCtx(s Style) renderCtx {
	return renderCtx{Style: s}
}

// displayValue returns v as it should appear in output for this style:
// asciified when ASCII mode is set, otherwise verbatim. Called by both
// measure and render so the width and the painted glyphs agree.
func (s renderCtx) displayValue(v string) string {
	if s.ASCII {
		return asciify(v)
	}
	return v
}

// ANSI escape sequences
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiItalic = "\033[3m"
)

func ansiColor(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// Semantic role tags. Each renderer call site asks for the color of a
// role rather than reaching into the palette directly — keeps the
// "color when Color is on, empty string when off" check in one place.
type semRole int

const (
	semVariable semRole = iota
	semNumber
	semOperator
	semText
	semDelim
	semBar
	semBigOp
)

var semColors = [...]string{
	semVariable: ansiColor(210, 180, 140), // warm tan
	semNumber:   ansiColor(255, 200, 120), // gold
	semOperator: ansiColor(180, 180, 220), // soft blue-gray
	semText:     ansiColor(200, 200, 200), // light gray
	semDelim:    ansiColor(140, 160, 190), // steel blue
	semBar:      ansiColor(100, 120, 150), // muted blue
	semBigOp:    ansiColor(180, 140, 200), // soft purple
}

// color returns the ANSI escape for r when Color is enabled, or the
// empty string when it isn't. Render sites pass the result straight
// through to setColored/putStrColored without the usual if-guard.
func (s renderCtx) color(r semRole) string {
	if !s.Color {
		return ""
	}
	return semColors[r]
}

// glyphs holds the character set used for box drawing, delimiters, and
// other layout primitives. Two variants exist: Unicode (default) and ASCII.
type glyphs struct {
	FracBar  rune // horizontal fraction bar
	OverBar  rune // bar above (\overline, sqrt extension)
	UnderBar rune
	Hat      rune
	Sqrt     rune

	VBar       rune // single vertical bar (|)
	VBarDouble rune // double vertical bar (‖)

	// Tall-delimiter parts: top, middle (extension), bottom.
	ParenLT, ParenLM, ParenLB          rune // ⎛ ⎜ ⎝
	ParenRT, ParenRM, ParenRB          rune // ⎞ ⎟ ⎠
	BrackLT, BrackLM, BrackLB          rune // ⎡ ⎢ ⎣
	BrackRT, BrackRM, BrackRB          rune // ⎤ ⎥ ⎦
	BraceLT, BraceLM, BraceLE, BraceLB rune // ⎧ ⎨ ⎪ ⎩
	BraceRT, BraceRM, BraceRE, BraceRB rune // ⎫ ⎬ ⎪ ⎭

	// Single-cell accent marks for \hat, \dot, \ddot, \tilde, \vec.
	HatMark, DotMark, DDotMark, TildeMark, VecMark rune

	// Overbrace / underbrace decorations.
	OverbraceLeft, OverbraceMid, OverbraceRight    rune // ╭ ┴ ╮
	UnderbraceLeft, UnderbraceMid, UnderbraceRight rune // ╰ ┬ ╯
}

var unicodeGlyphs = glyphs{
	FracBar:  '─',
	OverBar:  '‾',
	UnderBar: '_',
	Hat:      '^',
	Sqrt:     '√',

	VBar:       '│',
	VBarDouble: '‖',

	ParenLT: '⎛', ParenLM: '⎜', ParenLB: '⎝',
	ParenRT: '⎞', ParenRM: '⎟', ParenRB: '⎠',
	BrackLT: '⎡', BrackLM: '⎢', BrackLB: '⎣',
	BrackRT: '⎤', BrackRM: '⎥', BrackRB: '⎦',
	BraceLT: '⎧', BraceLM: '⎨', BraceLE: '⎪', BraceLB: '⎩',
	BraceRT: '⎫', BraceRM: '⎬', BraceRE: '⎪', BraceRB: '⎭',

	HatMark:   '^',
	DotMark:   '˙',
	DDotMark:  '¨',
	TildeMark: '~',
	VecMark:   '→',

	OverbraceLeft:   '╭',
	OverbraceMid:    '┴',
	OverbraceRight:  '╮',
	UnderbraceLeft:  '╰',
	UnderbraceMid:   '┬',
	UnderbraceRight: '╯',
}

var asciiGlyphs = glyphs{
	FracBar:  '-',
	OverBar:  '_',
	UnderBar: '_',
	Hat:      '^',
	Sqrt:     '\\',

	VBar:       '|',
	VBarDouble: '|',

	ParenLT: '(', ParenLM: '|', ParenLB: '(',
	ParenRT: ')', ParenRM: '|', ParenRB: ')',
	BrackLT: '[', BrackLM: '|', BrackLB: '[',
	BrackRT: ']', BrackRM: '|', BrackRB: ']',
	BraceLT: '{', BraceLM: '<', BraceLE: '|', BraceLB: '{',
	BraceRT: '}', BraceRM: '>', BraceRE: '|', BraceRB: '}',

	HatMark:   '^',
	DotMark:   '.',
	DDotMark:  ':',
	TildeMark: '~',
	VecMark:   '>',

	OverbraceLeft:   '+',
	OverbraceMid:    '^',
	OverbraceRight:  '+',
	UnderbraceLeft:  '+',
	UnderbraceMid:   'v',
	UnderbraceRight: '+',
}

// glyphs returns the glyph set for this style.
func (s Style) glyphs() glyphs {
	if s.ASCII {
		return asciiGlyphs
	}
	return unicodeGlyphs
}
