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
	// Inline typesets in TeX's text style rather than display style:
	// fractions flatten to a/b, big-operator limits move to the side
	// (∑ᵢ₌₁ⁿ instead of stacked), so simple expressions stay on one
	// row and can sit inside a sentence. Use it for $...$ in markdown;
	// leave it off for $$...$$.
	Inline bool
	// Width, when positive, is the maximum number of columns. An
	// expression wider than this is broken into several rows, before
	// relations (=, ≤, →) first and binary operators second, with
	// continuation rows indented. Zero means never break.
	Width int
}

// texStyle is TeX's notion of the current size context. Display and
// text differ in how fractions and limits are laid out; script and
// scriptscript additionally drop the optional inter-atom spaces so
// `i=1` stays tight inside a limit.
type texStyle uint8

const (
	styleDisplay texStyle = iota
	styleText
	styleScript
	styleScriptScript
)

// script returns the style used for sub/superscripts of this style.
func (m texStyle) script() texStyle {
	if m >= styleScript {
		return styleScriptScript
	}
	return styleScript
}

// renderCtx bundles user-facing [Style] options with derived runtime
// state used during a single measure-and-render pass. Keeping these
// separate from [Style] avoids exposing pass-internal flags on the
// public API.
type renderCtx struct {
	Style
	// style is the current math style. Render starts in display (or
	// text, for Style.Inline) and the script paths step it down.
	style texStyle
}

// withScript returns a copy of the context in the script style of the
// current one — used for sub/superscripts and big-operator limits.
func (c renderCtx) withScript() renderCtx {
	c.style = c.style.script()
	return c
}

// withStyle returns a copy of the context in the given style.
func (c renderCtx) withStyle(st texStyle) renderCtx {
	c.style = st
	return c
}

// styleSwitches maps the TeX style-switch commands to the style they
// select for the remainder of the enclosing list.
var styleSwitches = map[string]texStyle{
	"displaystyle":      styleDisplay,
	"textstyle":         styleText,
	"scriptstyle":       styleScript,
	"scriptscriptstyle": styleScriptScript,
}

// compact reports whether optional inter-atom spacing is suppressed
// (script and scriptscript styles).
func (c renderCtx) compact() bool {
	return c.style >= styleScript
}

// newRenderCtx wraps a [Style] in a fresh render context with no
// derived state set. Measurement memoization lives on the AST node.
func newRenderCtx(s Style) renderCtx {
	ctx := renderCtx{Style: s}
	if s.Inline {
		ctx.style = styleText
	}
	return ctx
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

	// Single-cell accent marks for \hat, \dot, \ddot, \tilde, \vec,
	// \breve, \check, \acute, \grave, \mathring.
	HatMark, DotMark, DDotMark, TildeMark, VecMark       rune
	BreveMark, CheckMark, AcuteMark, GraveMark, RingMark rune

	// Arrow heads for \overrightarrow / \overleftarrow.
	ArrowLeft, ArrowRight rune

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
	BreveMark: '˘',
	CheckMark: 'ˇ',
	AcuteMark: '´',
	GraveMark: '`',
	RingMark:  '˚',

	ArrowLeft:  '←',
	ArrowRight: '→',

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
	BreveMark: 'u',
	CheckMark: 'v',
	AcuteMark: '\'',
	GraveMark: '`',
	RingMark:  'o',

	ArrowLeft:  '<',
	ArrowRight: '>',

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
