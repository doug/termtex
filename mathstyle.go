package termtex

import "strings"

// mathStyle encodes a single Mathematical Alphanumeric Symbols block.
// upper/lower/digit are the base codepoints (0 = "not supported, pass
// through unchanged"). overrides redirects specific letters to the
// canonical BMP letterlike codepoints that occupy the reserved holes
// in the block.
type mathStyle struct {
	upper, lower, digit rune
	overrides           map[rune]rune
	// Greek bases (0 = unsupported): the Mathematical Greek blocks are
	// laid out as 25 capitals (Α…Ω with ϴ in the final-sigma slot),
	// ∇, 25 lowercase (α…ω including ς), ∂, then ϵ ϑ ϰ ϕ ϱ ϖ.
	gUpper, gLower rune
}

// greekBlockSize is the stride between styled Greek blocks.
const greekBlockSize = 58

// mathGreekToPlain maps a styled Greek letter (bold, italic, …) back to
// its plain codepoint, or 0 if r is not in the styled Greek range.
func mathGreekToPlain(r rune) rune {
	if r < 0x1D6A8 || r > 0x1D7CB {
		return 0
	}
	idx := int(r-0x1D6A8) % greekBlockSize
	switch {
	case idx < 25:
		return 0x391 + rune(idx)
	case idx == 25:
		return '∇'
	case idx < 51:
		return 0x3B1 + rune(idx-26)
	case idx == 51:
		return '∂'
	}
	return []rune("ϵϑϰϕϱϖ")[idx-52]
}

var (
	doubleStruckOverrides = map[rune]rune{
		'C': 0x2102, 'H': 0x210D, 'N': 0x2115, 'P': 0x2119,
		'Q': 0x211A, 'R': 0x211D, 'Z': 0x2124,
	}
	scriptOverrides = map[rune]rune{
		'B': 0x212C, 'E': 0x2130, 'F': 0x2131, 'H': 0x210B,
		'I': 0x2110, 'L': 0x2112, 'M': 0x2133, 'R': 0x211B,
		'e': 0x212F, 'g': 0x210A, 'o': 0x2134,
	}
	frakturOverrides = map[rune]rune{
		'C': 0x212D, 'H': 0x210C, 'I': 0x2111, 'R': 0x211C, 'Z': 0x2128,
	}
	// Italic lowercase 'h' has no codepoint at U+1D455 (reserved);
	// the canonical italic h lives at U+210E.
	italicOverrides = map[rune]rune{'h': 0x210E}
)

var mathStyles = map[string]mathStyle{
	"bold":         {0x1D400, 0x1D41A, 0x1D7CE, nil, 0x1D6A8, 0x1D6C2},
	"italic":       {0x1D434, 0x1D44E, 0, italicOverrides, 0x1D6E2, 0x1D6FC},
	"doubleStruck": {0x1D538, 0x1D552, 0x1D7D8, doubleStruckOverrides, 0, 0},
	"script":       {0x1D49C, 0x1D4B6, 0, scriptOverrides, 0, 0},
	"fraktur":      {0x1D504, 0x1D51E, 0, frakturOverrides, 0, 0},
	"sansSerif":    {0x1D5A0, 0x1D5BA, 0x1D7E2, nil, 0, 0},
	"monospace":    {0x1D670, 0x1D68A, 0x1D7F6, nil, 0, 0},
}

// mathCancel and mathStrike overlay a combining long solidus (\cancel)
// or long stroke (\sout) on every rune; both are zero-width in the
// grid so the box does not change.
func mathCancel(s string) string { return overlay(s, '̸') }
func mathStrike(s string) string { return overlay(s, '̶') }

func overlay(s string, mark rune) string {
	var sb strings.Builder
	sb.Grow(len(s) * 3)
	for _, r := range s {
		sb.WriteRune(r)
		sb.WriteRune(mark)
	}
	return sb.String()
}

// transform applies the style to s, emitting the styled codepoint for
// each A-Z/a-z/0-9 (when a base is set) and passing other runes through.
func (st mathStyle) transform(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if o, ok := st.overrides[r]; ok {
			sb.WriteRune(o)
			continue
		}
		switch {
		case st.upper != 0 && r >= 'A' && r <= 'Z':
			sb.WriteRune(st.upper + (r - 'A'))
		case st.lower != 0 && r >= 'a' && r <= 'z':
			sb.WriteRune(st.lower + (r - 'a'))
		case st.digit != 0 && r >= '0' && r <= '9':
			sb.WriteRune(st.digit + (r - '0'))
		case st.gUpper != 0 && r >= 0x391 && r <= 0x3A9 && r != 0x3A2:
			sb.WriteRune(st.gUpper + (r - 0x391))
		case st.gLower != 0 && r >= 0x3B1 && r <= 0x3C9:
			sb.WriteRune(st.gLower + (r - 0x3B1))
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func mathBold(s string) string         { return mathStyles["bold"].transform(s) }
func mathItalic(s string) string       { return mathStyles["italic"].transform(s) }
func mathDoubleStruck(s string) string { return mathStyles["doubleStruck"].transform(s) }
func mathScript(s string) string       { return mathStyles["script"].transform(s) }
func mathFraktur(s string) string      { return mathStyles["fraktur"].transform(s) }
func mathSansSerif(s string) string    { return mathStyles["sansSerif"].transform(s) }
func mathMonospace(s string) string    { return mathStyles["monospace"].transform(s) }

// applyMathStyle walks n in place, applying transform to the Value of
// every nodeSymbol and nodeNumber. The caller is parseMathStyle, which
// just parsed the subtree — nothing else references it yet, so mutation
// is safe and avoids a full deep copy on every \mathbb{...}/\mathbf{...}.
func applyMathStyle(n *node, transform func(string) string) *node {
	if n == nil {
		return nil
	}
	if n.Type == nodeSymbol || n.Type == nodeNumber {
		n.Value = transform(n.Value)
	}
	for _, c := range n.Children {
		applyMathStyle(c, transform)
	}
	for _, row := range n.Rows {
		for _, c := range row {
			applyMathStyle(c, transform)
		}
	}
	return n
}
