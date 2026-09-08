package termtex

// Unicode superscript and subscript mappings for inline rendering.
// When an exponent or subscript is a single character with a Unicode
// equivalent, we render it inline (1 row) instead of stacking.

var superscriptMap = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼',
	'(': '⁽', ')': '⁾',
	'n': 'ⁿ', 'i': 'ⁱ',
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'j': 'ʲ', 'k': 'ᵏ',
	'l': 'ˡ', 'm': 'ᵐ', 'o': 'ᵒ', 'p': 'ᵖ', 'r': 'ʳ',
	's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ', 'v': 'ᵛ', 'w': 'ʷ',
	'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
	// uppercase
	'A': 'ᴬ', 'B': 'ᴮ', 'D': 'ᴰ', 'E': 'ᴱ', 'G': 'ᴳ',
	'H': 'ᴴ', 'I': 'ᴵ', 'J': 'ᴶ', 'K': 'ᴷ', 'L': 'ᴸ',
	'M': 'ᴹ', 'N': 'ᴺ', 'O': 'ᴼ', 'P': 'ᴾ', 'R': 'ᴿ',
	'T': 'ᵀ', 'U': 'ᵁ', 'V': 'ⱽ', 'W': 'ᵂ',
	// Greek
	'α': 'ᵅ', 'β': 'ᵝ', 'γ': 'ᵞ', 'δ': 'ᵟ', 'θ': 'ᶿ',
	'φ': 'ᵠ', 'χ': 'ᵡ',
	// special
	'′': '′', // prime is already a superscript glyph
}

var subscriptMap = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌',
	'(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
	'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
	'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
	'v': 'ᵥ', 'x': 'ₓ',
	// Greek
	'β': 'ᵦ', 'γ': 'ᵧ', 'ρ': 'ᵨ', 'φ': 'ᵩ', 'χ': 'ᵪ',
}

// canInlineScript reports whether n can be rendered as a sequence of
// inline Unicode codepoints (single line, no stacking) using the given
// rune map. Only flat runs of symbols, numbers and operators qualify;
// a script that itself carries a script (`x^{y^z}`) stacks, since the
// flattened form `xʸᶻ` would not show the nesting. ASCII mode never
// inlines.
//
// Each script decides for itself: `T_c + T_h` renders `T` with a
// stacked `c` next to an inline `Tₕ`, rather than forcing every
// script in the expression to stack because one letter has no
// subscript form.
func canInlineScript(n *node, s renderCtx, m map[rune]rune) bool {
	if n == nil || s.ASCII {
		return false
	}
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return allMapped(n.Value, m)
	case nodeGroup:
		if len(n.Children) == 0 {
			return false
		}
		for _, ch := range n.Children {
			if !canInlineScript(ch, s, m) {
				return false
			}
		}
		return true
	}
	return false
}

func canInlineSuperscript(n *node, s renderCtx) bool {
	return canInlineScript(n, s, superscriptMap)
}

func canInlineSubscript(n *node, s renderCtx) bool {
	return canInlineScript(n, s, subscriptMap)
}

func allMapped(s string, m map[rune]rune) bool {
	for _, r := range s {
		if _, ok := m[r]; !ok {
			return false
		}
	}
	return len(s) > 0
}

// toScript walks n applying m to leaf string values.
func toScript(n *node, m map[rune]rune) string {
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return mapRunes(n.Value, m)
	case nodeGroup:
		var s string
		for _, ch := range n.Children {
			s += toScript(ch, m)
		}
		return s
	}
	return ""
}

func toSuperscript(n *node) string { return toScript(n, superscriptMap) }
func toSubscript(n *node) string   { return toScript(n, subscriptMap) }

func mapRunes(s string, m map[rune]rune) string {
	var out []rune
	for _, r := range s {
		if mapped, ok := m[r]; ok {
			out = append(out, mapped)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// inlineScriptWidth returns the display width of the inline superscript/subscript form.
func inlineScriptWidth(n *node) int {
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return displayWidth(n.Value)
	case nodeGroup:
		w := 0
		for _, ch := range n.Children {
			w += inlineScriptWidth(ch)
		}
		return w
	case nodeScript:
		base, sub, sup := scriptParts(n)
		w := inlineScriptWidth(base)
		if sup != nil {
			w += inlineScriptWidth(sup)
		}
		if sub != nil {
			w += inlineScriptWidth(sub)
		}
		return w
	}
	return 0
}
