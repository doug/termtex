package termtex

// Inter-atom spacing modeled on TeX's math lists (TeXbook, chapter 18).
// Every atom carries a class — Ord, Op, Bin, Rel, Open, Close, Punct or
// Inner — and the gap between two neighbours is a pure function of
// their classes and the current math style. On a character grid the
// thin/medium/thick distinction collapses to "one cell or none", but
// the *structure* of the table is what makes `x = 1, 2, …, n`,
// `det(A)`, `sin x` and `∫ f(x) dx` all come out right from one rule.

type atomClass uint8

const (
	atomUnset atomClass = iota // zero value: derive from node type
	atomOrd                    // ordinary symbol: x, 2, α
	atomOp                     // large/named operator: ∑, ∫, sin, lim
	atomBin                    // binary operator: +, ×, ∪
	atomRel                    // relation: =, <, ∈, →
	atomOpen                   // opening delimiter: (, [, ⟨
	atomClose                  // closing delimiter: ), ], ⟩, !
	atomPunct                  // punctuation: , ;
	atomInner                  // fraction-like inner list
	atomNone                   // transparent (explicit spaces)
)

// Glyph → class tables for operator and symbol nodes. Anything not
// listed is Ord, which is also what TeX does for unknown characters.
var (
	binGlyphs = runeSet("+-−±∓×÷·∗*∘•⋆∖∪∩⊕⊗⊖⊘⊙⊚⊛⊝⊞⊟⊠⊡∧∨∙⊔⊓⊎⨿≀◯△▽◁▷⊲⊳⊴⊵†‡⋄⋅⋉⋊⋋⋌⋎⋏⊻⊼⌆⋇∔⊺")
	relGlyphs = runeSet("=<>≤≥≠≈≡∼≃≅≇≁∝≐≑≜≊≍⋈≖≗≓≒∽≂" +
		"∈∉∋∌⊂⊆⊃⊇⊄⊅⊈⊉⊊⊋⋐⋑⊏⊐⊑⊒" +
		"→←↦⇒⇐↔⇔⟹⟸⟺⟶⟵⟷⟼↪↩↠↞⇀⇁↼↽⇌⇋⇄⇆⇉⇇↑↓↕⇑⇓⇕↗↘↙↖⇝↭↛↚⇏⇍↮⇎↺↻↶↷↰↱⇢⇠" +
		"⊥∥∦∣∤:≪≫⋘⋙⊢⊣⊩⊪⊬⊭⊨≺≻⪯⪰⊀⊁≾≿≔≕≮≯≰≱≦≧⩽⩾≲≳≶≷⪅⪆⋚⋛≬⋔⌣⌢∴∵≢≄≉∄")
	punctGlyphs = runeSet(",;")
	openGlyphs  = runeSet("([{⟨⌊⌈⌜⌞")
	closeGlyphs = runeSet(")]}⟩⌋⌉⌝⌟!")
	innerGlyphs = runeSet("…⋯⋮⋱⋰")
)

// runeSet builds a glyph lookup from a string of single-rune glyphs.
func runeSet(s string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, r := range s {
		m[string(r)] = true
	}
	return m
}

func glyphClass(v string) atomClass {
	// A negated relation built with a combining overlay keeps the
	// class of its base glyph.
	if len(v) > 2 && (v[len(v)-2:] == "̸") {
		v = v[:len(v)-2]
	}
	switch {
	case binGlyphs[v]:
		return atomBin
	case relGlyphs[v]:
		return atomRel
	case punctGlyphs[v]:
		return atomPunct
	case openGlyphs[v]:
		return atomOpen
	case closeGlyphs[v]:
		return atomClose
	case innerGlyphs[v]:
		return atomInner
	}
	return atomOrd
}

// leftClass and rightClass report the atom class visible at each edge
// of a node. Most nodes are a single atom, so both edges agree; groups
// and delimited lists expose the class of their first/last element.
func leftClass(n *node) atomClass {
	if n == nil {
		return atomNone
	}
	if n.class != atomUnset {
		return n.class
	}
	switch n.Type {
	case nodeOperator, nodeSymbol:
		return glyphClass(n.Value)
	case nodeNumber, nodeText, nodeSqrt, nodeNthRoot, nodeMatrix:
		return atomOrd
	case nodeSpace:
		return atomNone
	case nodeBigOp, nodeLim:
		return atomOp
	case nodeXArrow:
		return atomRel
	case nodeFrac:
		return atomInner
	case nodeParen:
		if n.Open == "" {
			return leftClass(n.Children[0])
		}
		return atomOpen
	case nodeScript:
		return leftClass(n.Children[0])
	case nodeGroup:
		for _, ch := range n.Children {
			c := leftClass(ch)
			if c == atomNone {
				continue
			}
			if c == atomBin {
				return atomOrd // a leading binary operator is unary
			}
			return c
		}
		return atomNone
	case nodeOverline, nodeUnderline, nodeHat, nodeOverbrace, nodeUnderbrace, nodeStyle:
		return leftClass(n.Children[0])
	}
	return atomOrd
}

func rightClass(n *node) atomClass {
	if n == nil {
		return atomNone
	}
	if n.class != atomUnset {
		return n.class
	}
	switch n.Type {
	case nodeParen:
		if n.Close == "" {
			return rightClass(n.Children[0])
		}
		return atomClose
	case nodeScript:
		return rightClass(n.Children[0])
	case nodeGroup:
		for i := len(n.Children) - 1; i >= 0; i-- {
			c := rightClass(n.Children[i])
			if c == atomNone {
				continue
			}
			if c == atomBin {
				return atomOrd // a trailing binary operator has no right operand
			}
			return c
		}
		return atomNone
	case nodeOverline, nodeUnderline, nodeHat, nodeOverbrace, nodeUnderbrace, nodeStyle:
		return rightClass(n.Children[0])
	}
	return leftClass(n)
}

// spacingTable[left][right] gives the gap between two atoms: 0 none,
// 1 always one cell, 2 one cell except in script styles. Rows and
// columns are indexed by atomClass-1 (Ord … Inner).
//
// This follows TeX's table except that Punct never receives a space
// on its left (TeX uses a thin space before a comma after Inner or
// Punct; a full cell there reads as a typo on a grid).
var spacingTable = [8][8]uint8{
	//         Ord Op Bin Rel Open Close Punct Inner
	/* Ord   */ {0, 1, 2, 2, 0, 0, 0, 2},
	/* Op    */ {1, 1, 0, 2, 0, 0, 0, 2},
	/* Bin   */ {2, 2, 0, 0, 2, 0, 0, 2},
	/* Rel   */ {2, 2, 0, 0, 2, 0, 0, 2},
	/* Open  */ {0, 0, 0, 0, 0, 0, 0, 0},
	/* Close */ {0, 1, 2, 2, 0, 0, 0, 2},
	/* Punct */ {2, 2, 0, 2, 2, 0, 0, 2},
	/* Inner */ {2, 1, 2, 2, 2, 0, 0, 2},
}

func classGap(left, right atomClass, s renderCtx) int {
	if left == atomNone || right == atomNone || left == atomUnset || right == atomUnset {
		return 0
	}
	switch spacingTable[left-1][right-1] {
	case 1:
		return 1
	case 2:
		if s.compact() {
			return 0
		}
		return 1
	}
	return 0
}

// groupGaps returns, for each child of a horizontal list, the number of
// blank cells to insert before it. Both measureGroup and renderGroup
// call this so widths and painted positions agree.
//
// Binary operators are demoted to Ord when they have no left operand
// (start of list, or after Bin/Op/Rel/Open/Punct) or no right operand
// (before Rel/Close/Punct) — TeX's rule that turns `x = -1` and `(-b)`
// into unary minus.
func groupGaps(children []*node, s renderCtx) []int {
	n := len(children)
	gaps := make([]int, n)
	lc, rc, unary := effectiveClasses(children)

	prev := -1 // index of the previous non-transparent child
	for i, ch := range children {
		if lc[i] == atomNone {
			continue
		}
		if prev >= 0 {
			gaps[i] = classGap(rc[prev], lc[i], s)
			switch {
			case unary[prev]:
				// A unary sign hugs its operand even when the operand
				// is a named function: `-sin θ`, not `- sin θ`.
				gaps[i] = 0
			case gaps[i] == 0 && hasStackedLimits(children[prev], s):
				// Limits stacked under/over an operator are usually
				// wider than the symbol; give the next atom a cell so
				// `lim(` does not read as one glyph.
				gaps[i] = 1
			}
			if gaps[i] == 0 && needsBarSeparator(children[prev], ch, s) {
				gaps[i] = 1
			}
		}
		prev = i
	}
	return gaps
}

// effectiveClasses returns each child's left and right atom class
// after TeX's demotion rules, and which children are binary operators
// demoted to unary signs.
func effectiveClasses(children []*node) (lc, rc []atomClass, unary []bool) {
	n := len(children)
	lc = make([]atomClass, n)
	rc = make([]atomClass, n)
	unary = make([]bool, n)
	for i, ch := range children {
		lc[i] = leftClass(ch)
		rc[i] = rightClass(ch)
	}

	prev := -1 // index of the previous non-transparent child
	for i := range children {
		if lc[i] == atomNone {
			continue
		}
		prevR := atomNone
		if prev >= 0 {
			prevR = rc[prev]
		}
		if lc[i] == atomBin {
			switch prevR {
			case atomNone, atomBin, atomOp, atomRel, atomOpen, atomPunct:
				lc[i] = atomOrd
				if rc[i] == atomBin {
					rc[i] = atomOrd
					unary[i] = true
				}
			}
		}
		if prev >= 0 && rc[prev] == atomBin {
			switch lc[i] {
			case atomRel, atomClose, atomPunct:
				rc[prev] = atomOrd
				if lc[prev] == atomBin {
					lc[prev] = atomOrd
				}
			}
		}
		prev = i
	}
	return lc, rc, unary
}

// hasStackedLimits reports whether n is a big operator carrying
// limits laid out above/below it.
func hasStackedLimits(n *node, s renderCtx) bool {
	if n == nil || n.Type != nodeScript {
		return false
	}
	base, sub, sup := scriptParts(n)
	return isBigOp(base) && (sub != nil || sup != nil) && stacksScripts(base, s)
}
