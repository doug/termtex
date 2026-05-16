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
// rune map. isSuper picks the axis: true for superscript (allowing
// `x^{y^z}` to flatten), false for subscript. ASCII mode and the
// global stack override both force false.
func canInlineScript(n *node, s renderCtx, m map[rune]rune, isSuper bool) bool {
	if n == nil || s.ASCII || s.forceStackScripts {
		return false
	}
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return allMapped(n.Value, m)
	case nodeGroup:
		for _, ch := range n.Children {
			if !canInlineScript(ch, s, m, isSuper) {
				return false
			}
		}
		return true
	case nodeScript:
		base, sub, sup := scriptParts(n)
		if isSuper {
			return sub == nil && sup != nil &&
				canInlineScript(base, s, m, isSuper) &&
				canInlineScript(sup, s, m, isSuper)
		}
		return sup == nil && sub != nil &&
			canInlineScript(base, s, m, isSuper) &&
			canInlineScript(sub, s, m, isSuper)
	}
	return false
}

func canInlineSuperscript(n *node, s renderCtx) bool {
	return canInlineScript(n, s, superscriptMap, true)
}

func canInlineSubscript(n *node, s renderCtx) bool {
	return canInlineScript(n, s, subscriptMap, false)
}

func allMapped(s string, m map[rune]rune) bool {
	for _, r := range s {
		if _, ok := m[r]; !ok {
			return false
		}
	}
	return len(s) > 0
}

// toScript walks n applying m to leaf string values; nested nodeScript
// nodes along the same axis (sup for isSuper=true, sub otherwise) are
// flattened by concatenating base and script.
func toScript(n *node, m map[rune]rune, isSuper bool) string {
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return mapRunes(n.Value, m)
	case nodeGroup:
		var s string
		for _, ch := range n.Children {
			s += toScript(ch, m, isSuper)
		}
		return s
	case nodeScript:
		base, sub, sup := scriptParts(n)
		if isSuper && sup != nil {
			return toScript(base, m, isSuper) + toScript(sup, m, isSuper)
		}
		if !isSuper && sub != nil {
			return toScript(base, m, isSuper) + toScript(sub, m, isSuper)
		}
	}
	return ""
}

func toSuperscript(n *node) string { return toScript(n, superscriptMap, true) }
func toSubscript(n *node) string   { return toScript(n, subscriptMap, false) }

// canInlineSupRaw is canInlineSuperscript ignoring forceStackScripts.
// Used by callers (like nth-root indices) that always render via the
// inline path regardless of the script consistency rule.
func canInlineSupRaw(n *node, s renderCtx) bool {
	s.forceStackScripts = false
	return canInlineSuperscript(n, s)
}

// isSimpleScript reports whether a sub/superscript content is a single
// atom or group of atoms — the kind that has a chance of inlining as
// Unicode super/subscript characters. Complex scripts like fractions,
// roots, or matrices are not "simple" and always stack regardless.
func isSimpleScript(n *node) bool {
	if n == nil {
		return false
	}
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator:
		return true
	case nodeGroup, nodeScript:
		for _, ch := range n.Children {
			if ch != nil && !isSimpleScript(ch) {
				return false
			}
		}
		return true
	}
	return false
}

// hasMixedSimpleScripts walks the AST and returns true if any "simple"
// sub/superscript can't be rendered inline as a Unicode codepoint
// (e.g. `T_c` — `c` has no Unicode subscript). When true, all simple
// scripts are forced to stack so the rendering is uniform across the
// expression rather than mixing `Tₕ` (inline) with stacked `T \n c`.
func hasMixedSimpleScripts(n *node, s renderCtx) bool {
	test := s
	test.forceStackScripts = false

	var walk func(*node) bool
	walk = func(node *node) bool {
		if node == nil {
			return false
		}
		if node.Type == nodeScript {
			base, sub, sup := scriptParts(node)
			if !isBigOp(base) {
				if sub != nil && isSimpleScript(sub) && !canInlineSubscript(sub, test) {
					return true
				}
				if sup != nil && isSimpleScript(sup) && !canInlineSuperscript(sup, test) {
					return true
				}
			}
		}
		for _, ch := range node.Children {
			if ch != nil && walk(ch) {
				return true
			}
		}
		for _, row := range node.Rows {
			for _, c := range row {
				if walk(c) {
					return true
				}
			}
		}
		return false
	}
	return walk(n)
}

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
