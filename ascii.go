package termtex

import "strings"

// asciiRuneMap maps Unicode math characters to ASCII strings.
// Keys are runes that termtex emits during parsing (Greek letters,
// math operators, big operators, arrows, etc.) and the values are
// their pure-ASCII equivalents.
var asciiRuneMap = map[rune]string{
	// Lowercase Greek
	'α': "alpha", 'β': "beta", 'γ': "gamma", 'δ': "delta",
	'ε': "epsilon", 'ζ': "zeta", 'η': "eta", 'θ': "theta",
	'ι': "iota", 'κ': "kappa", 'λ': "lambda", 'μ': "mu",
	'ν': "nu", 'ξ': "xi", 'ο': "o", 'π': "pi",
	'ρ': "rho", 'σ': "sigma", 'τ': "tau", 'υ': "upsilon",
	'φ': "phi", 'χ': "chi", 'ψ': "psi", 'ω': "omega",

	// Uppercase Greek
	'Γ': "Gamma", 'Δ': "Delta", 'Θ': "Theta", 'Λ': "Lambda",
	'Ξ': "Xi", 'Π': "Pi", 'Σ': "Sigma", 'Φ': "Phi",
	'Ψ': "Psi", 'Ω': "Omega",

	// Math operators
	'±': "+/-", '∓': "-/+", '×': "x", '÷': "/", '·': "*",
	'≤': "<=", '≥': ">=", '≠': "!=", '≈': "~=", '≡': "==",
	'∼': "~", '≅': "~=", '∝': "prop",
	'∗': "*", '∘': "o", '•': "*", '⋆': "*",

	// Set theory
	'∈': "in", '∉': "!in", '⊂': "sub", '⊃': "sup",
	'⊆': "sube", '⊇': "supe", '∪': "U", '∩': "^",
	'∅': "{}", '∖': "\\",

	// Logic
	'∀': "forall", '∃': "exists", '¬': "not",
	'∧': "&&", '∨': "||",

	// Arrows
	'→': "->", '←': "<-", '↦': "|->",
	'⇒': "=>", '⇐': "<=", '↔': "<->", '⇔': "<=>",

	// Brackets / dots
	'⟨': "<", '⟩': ">",
	'…': "...", '⋯': "...", '⋮': ":", '⋱': "\\",

	// Misc
	'∞': "inf", 'ℏ': "hbar", 'ℓ': "l",
	'∂': "d", '∇': "del", '′': "'",

	// Big operators
	'∑': "Sum", '∏': "Prod", '∫': "int", '∮': "oint",

	// Tall delimiters
	'‖': "|",

	// Box-drawing leftovers (defensive: should already be swapped via glyphs)
	'─': "-", '│': "|", '‾': "_", '√': "\\",
}

// letterlikeToASCII reverses the BMP letterlike codepoints used as
// holes in the Mathematical Alphanumeric Symbols block (script B,
// double-struck R, fraktur H, italic h, etc.) back to plain ASCII.
var letterlikeToASCII = map[rune]rune{
	// Italic h
	0x210E: 'h',
	// Script (mathcal) holes
	0x212C: 'B', 0x2130: 'E', 0x2131: 'F', 0x210B: 'H',
	0x2110: 'I', 0x2112: 'L', 0x2133: 'M', 0x211B: 'R',
	0x212F: 'e', 0x210A: 'g', 0x2134: 'o',
	// Fraktur holes
	0x212D: 'C', 0x210C: 'H', 0x2111: 'I', 0x211C: 'R', 0x2128: 'Z',
	// Double-struck holes
	0x2102: 'C', 0x210D: 'H', 0x2115: 'N', 0x2119: 'P',
	0x211A: 'Q', 0x211D: 'R', 0x2124: 'Z',
}

// mathAlphanumToASCII reverses a Mathematical Alphanumeric Symbols
// codepoint back to plain A-Z / a-z / 0-9, or returns 0 if r isn't
// in the styled-letter range.
func mathAlphanumToASCII(r rune) rune {
	if a, ok := letterlikeToASCII[r]; ok {
		return a
	}
	if r < 0x1D400 || r > 0x1D7FF {
		return 0
	}
	// Digits live at U+1D7CE..U+1D7FF (5 styles × 10 digits).
	if r >= 0x1D7CE {
		return rune('0' + int(r-0x1D7CE)%10)
	}
	// Letters live at U+1D400..U+1D6A3 organized in 52-letter style
	// blocks (26 uppercase, 26 lowercase). Greek styled letters
	// occupy U+1D6A4..U+1D7CD and aren't handled here — they fall
	// through to the asciiRuneMap or '?'.
	if r > 0x1D6A3 {
		return 0
	}
	idx := int(r-0x1D400) % 52
	if idx < 26 {
		return rune('A' + idx)
	}
	return rune('a' + (idx - 26))
}

// asciify rewrites a string so it contains only 7-bit ASCII.
// Runes with mappings in asciiRuneMap are replaced; runes already in
// ASCII range pass through; any other non-ASCII rune is replaced with '?'.
func asciify(s string) string {
	if s == "" {
		return s
	}
	allASCII := true
	for _, r := range s {
		if r > 0x7F {
			allASCII = false
			break
		}
	}
	if allASCII {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r <= 0x7F {
			b.WriteRune(r)
			continue
		}
		if a := mathAlphanumToASCII(r); a != 0 {
			b.WriteRune(a)
			continue
		}
		if mapped, ok := asciiRuneMap[r]; ok {
			b.WriteString(mapped)
			continue
		}
		b.WriteRune('?')
	}
	return b.String()
}
