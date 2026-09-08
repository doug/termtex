package termtex

import (
	"fmt"
	"unicode/utf8"
)

// box represents a measured rectangular region on the character grid.
type box struct {
	Width    int // horizontal cells
	Height   int // vertical cells
	Baseline int // row index (from top) where the main text line sits
}

// measure computes the [box] for an AST node, memoizing the result on
// the node itself so repeated walks (initial measure + render) do not
// recompute. One slot per math style.
func measure(n *node, s renderCtx) box {
	if n == nil {
		return box{Width: 0, Height: 1, Baseline: 0}
	}
	slot := uint8(s.style)
	mask := uint8(1) << slot
	if n.measured&mask != 0 {
		return n.boxes[slot]
	}
	b := measureUncached(n, s)
	n.boxes[slot] = b
	n.measured |= mask
	return b
}

// measureUncached is the bottom-up measurement function. Callers
// should go through [measure] so results are memoized.
func measureUncached(n *node, s renderCtx) box {
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator, nodeText:
		return box{Width: displayWidth(s.displayValue(n.Value)), Height: 1, Baseline: 0}
	case nodeSpace:
		if len(n.Children) > 0 {
			// \phantom: the size of its argument, painted as nothing.
			return measure(n.Children[0], s)
		}
		return box{Width: n.Width, Height: 1, Baseline: 0}
	case nodeGroup:
		return measureGroup(n, s)
	case nodeFrac:
		return measureFrac(n, s)
	case nodeScript:
		return measureScript(n, s)
	case nodeSqrt:
		return measureSqrt(n, s)
	case nodeNthRoot:
		return measureNthRoot(n, s)
	case nodeParen:
		return measureParen(n, s)
	case nodeMatrix:
		return measureMatrix(n, s)
	case nodeBigOp:
		return measureBigOp(n, s)
	case nodeLim:
		return box{Width: displayWidth(n.Value), Height: 1, Baseline: 0}
	case nodeXArrow:
		return measureXArrow(n, s)
	case nodeStyle:
		return measure(n.Children[0], s.withStyle(styleSwitches[n.Value]))
	case nodeOverline:
		return measureOverline(n, s)
	case nodeUnderline:
		return measureUnderline(n, s)
	case nodeHat:
		return measureHat(n, s)
	case nodeOverbrace:
		return measureOverbrace(n, s)
	case nodeUnderbrace:
		return measureUnderbrace(n, s)
	default:
		// Unknown node type — surface the bug rather than silently
		// rendering a 1x1 box.
		panic(fmt.Sprintf("termtex: measure: unhandled nodeType %d", n.Type))
	}
}

// needsBarSeparator returns true when prev and cur would render with
// adjacent horizontal strokes that visually merge — e.g. a unary `-`
// before a stacked fraction draws as `-────`, reading as one
// continuous line. Only fires when the spacing table gave no gap.
func needsBarSeparator(prev, cur *node, s renderCtx) bool {
	if prev == nil || cur == nil || prev.Type != nodeOperator {
		return false
	}
	switch prev.Value {
	case "-", "+", "−", "±", "∓":
	default:
		return false
	}
	return cur.Type == nodeFrac && !fracFlat(cur, s)
}

func measureGroup(n *node, s renderCtx) box {
	if len(n.Children) == 0 {
		// An empty group contributes no rows. Without this, constructs
		// that add label/index height (\overbrace, \sum^{}, \hat{}, …)
		// reserve a phantom blank row in the output.
		return box{Width: 0, Height: 0, Baseline: 0}
	}
	gaps := groupGaps(n.Children, s)
	totalWidth := 0
	maxAbove := 0 // rows above baseline
	maxBelow := 0 // rows below baseline (including baseline row)
	for i, child := range n.Children {
		b := measure(child, s)
		totalWidth += gaps[i] + b.Width
		above := b.Baseline
		below := b.Height - b.Baseline - 1
		if above > maxAbove {
			maxAbove = above
		}
		if below > maxBelow {
			maxBelow = below
		}
	}
	return box{
		Width:    totalWidth,
		Height:   maxAbove + 1 + maxBelow,
		Baseline: maxAbove,
	}
}

// fracFlat reports whether a fraction renders on one row as `a/b`
// rather than stacked over a bar. Display style always stacks; text
// and script styles flatten whenever both sides fit on one row.
// \dfrac and \tfrac override the style-driven choice.
func fracFlat(n *node, s renderCtx) bool {
	switch n.Value {
	case "d":
		return false
	case "t":
		// \tfrac asks for the flat form; it still needs single-row sides.
	default:
		if s.style == styleDisplay {
			return false
		}
	}
	return measure(n.Children[0], s).Height <= 1 && measure(n.Children[1], s).Height <= 1
}

// flatFracWrap reports whether a side of a flat fraction needs
// parentheses to stay unambiguous. Numerators only need them when they
// contain an operator (`(a + b)/2` but `n(n+1)/2`); denominators need
// them for any multi-atom content (`1/(2a)`).
func flatFracWrap(n *node, den bool) bool {
	if n == nil {
		return false
	}
	switch n.Type {
	case nodeFrac:
		return true
	case nodeGroup:
		if len(n.Children) <= 1 {
			return false
		}
		if den {
			return true
		}
		for _, ch := range n.Children {
			if ch.Type == nodeOperator {
				return true
			}
		}
	}
	return false
}

func measureFrac(n *node, s renderCtx) box {
	num := measure(n.Children[0], s)
	den := measure(n.Children[1], s)
	if fracFlat(n, s) {
		w := num.Width + 1 + den.Width
		if flatFracWrap(n.Children[0], false) {
			w += 2
		}
		if flatFracWrap(n.Children[1], true) {
			w += 2
		}
		return box{Width: w, Height: 1, Baseline: 0}
	}
	inner := max(num.Width, den.Width)
	// Minimal padding: just enough to visually separate the bar from neighbors
	pad := 0
	if inner <= 2 {
		pad = 1 // single-char fractions get 1 cell padding per side
	}
	w := inner + pad*2
	h := num.Height + 1 + den.Height
	return box{Width: w, Height: h, Baseline: num.Height}
}

func isBigOp(n *node) bool {
	return n.Type == nodeBigOp || n.Type == nodeLim
}

// intFamily lists the big operators whose limits sit to the side even
// in display style, as TeX does for integrals.
var intFamily = map[string]bool{"∫": true, "∬": true, "∭": true, "∮": true, "∯": true, "∰": true}

// limitsStacked decides where a big operator's limits go: stacked
// above and below (display-style ∑, ∏, lim) or to the side as ordinary
// scripts (integrals, and everything in text style). \limits and
// \nolimits override.
func limitsStacked(base *node, s renderCtx) bool {
	if base.limits > 0 {
		return true
	}
	if base.limits < 0 {
		return false
	}
	if intFamily[base.Value] {
		return false
	}
	return s.style == styleDisplay
}

// textualScripts reports whether a script node renders in TeX-like
// notation on one row: `T_c`, `x^π`, `lim_{n→∞}`. Used outside display
// style when a script cannot be inlined as Unicode but is itself a
// single row, so an expression inside a sentence never grows a second
// row for want of a subscript glyph.
func textualScripts(base *node, bb box, sub, sup *node, s renderCtx) bool {
	if s.style == styleDisplay || bb.Height != 1 || stacksScripts(base, s) {
		return false
	}
	cs := s.withScript()
	if sub != nil && measure(sub, cs).Height != 1 {
		return false
	}
	if sup != nil && measure(sup, cs).Height != 1 {
		return false
	}
	return true
}

// textualScriptWidth is the width of one script in textual form: its
// Unicode inline form when it has one, otherwise the marker plus the
// content, braced when longer than one cell.
func textualScriptWidth(n *node, inline bool, s renderCtx) int {
	if n == nil {
		return 0
	}
	if inline {
		return inlineScriptWidth(n)
	}
	w := measure(n, s.withScript()).Width
	if w <= 1 {
		return 1 + w
	}
	return 3 + w
}

// stacksScripts reports whether a script node lays its scripts out
// above/below the base: big operators per limitsStacked, and any base
// tagged by \overset / \underset.
func stacksScripts(base *node, s renderCtx) bool {
	if base == nil {
		return false
	}
	return (isBigOp(base) || base.limits == 2) && limitsStacked(base, s)
}

// measureXArrow: the labels sit above (and optionally below) an arrow
// stretched to cover the wider label, with one cell of margin.
func measureXArrow(n *node, s renderCtx) box {
	cs := s.withScript()
	above := measure(n.Children[0], cs)
	w := above.Width + 2
	h := above.Height + 1
	if w < 3 {
		w = 3
	}
	var below box
	if n.Children[1] != nil {
		below = measure(n.Children[1], cs)
		if below.Width+2 > w {
			w = below.Width + 2
		}
		h += below.Height
	}
	return box{Width: w, Height: h, Baseline: above.Height}
}

// scriptBaseWrapped reports whether a script's base must be wrapped in
// parentheses to keep the script unambiguous: a flat fraction, so
// `\frac{a}{b}^2` renders `(a/b)²` rather than `a/b²`.
func scriptBaseWrapped(base *node, s renderCtx) bool {
	return base != nil && base.Type == nodeFrac && fracFlat(base, s)
}

// measureScript handles base[_sub][^sup]. Either sub or sup may be
// nil, and there are three distinct shapes depending on the base:
//
//   - big operators with stacked limits (\sum, \lim in display style):
//     the limits sit above/below the operator.
//   - inline scripts (single-cell Unicode super/sub forms): rendered
//     to the right of the base on the same row.
//   - stacked scripts: super goes upper-right, sub goes lower-right.
func measureScript(n *node, s renderCtx) box {
	base, sub, sup := scriptParts(n)
	bb := measure(base, s)
	if scriptBaseWrapped(base, s) {
		bb.Width += 2
	}
	cs := s.withScript()

	if stacksScripts(base, s) {
		var subBox, supBox box
		if sub != nil {
			subBox = measure(sub, cs)
		}
		if sup != nil {
			supBox = measure(sup, cs)
		}
		w := max(bb.Width, max(subBox.Width, supBox.Width))
		return box{
			Width:    w,
			Height:   supBox.Height + bb.Height + subBox.Height,
			Baseline: supBox.Height + bb.Baseline,
		}
	}

	// Inline shortcut requires a single-row base (see renderScript).
	supInline := bb.Height == 1 && sup != nil && canInlineSuperscript(sup, s)
	subInline := bb.Height == 1 && sub != nil && canInlineSubscript(sub, s)
	supInlineAlone := sup != nil && sub == nil && supInline
	subInlineAlone := sub != nil && sup == nil && subInline
	bothInline := sup != nil && sub != nil && supInline && subInline

	if supInlineAlone || subInlineAlone || bothInline {
		w := bb.Width
		if sup != nil && supInline {
			w += inlineScriptWidth(sup)
		}
		if sub != nil && subInline {
			w += inlineScriptWidth(sub)
		}
		return box{Width: w, Height: bb.Height, Baseline: bb.Baseline}
	}

	if textualScripts(base, bb, sub, sup, s) {
		w := bb.Width + textualScriptWidth(sub, subInline, s) + textualScriptWidth(sup, supInline, s)
		return box{Width: w, Height: 1, Baseline: 0}
	}

	// Stacked. Compute super/sub boxes only if present.
	var supBox, subBox box
	if sup != nil {
		supBox = measure(sup, cs)
	}
	if sub != nil {
		subBox = measure(sub, cs)
	}
	scriptW := max(subBox.Width, supBox.Width)
	w := bb.Width + scriptW

	switch {
	case sup != nil && sub != nil:
		// Stacked sub+sup: sup sits in its own rows above the base.
		// Sub's top row coincides with the base's last row (different
		// column, so no visual collision) only when the base has more
		// than one row — for a height-1 base there's no "last row"
		// distinct from the first, so sub takes its own row below.
		overlap := 0
		if bb.Height > 1 {
			overlap = 1
		}
		// Rows below the sup: the base, or the base minus the shared
		// row plus the sub, whichever reaches further (an empty sub
		// must not shrink the base).
		below := max(bb.Height, bb.Height-overlap+subBox.Height)
		return box{
			Width:    w,
			Height:   supBox.Height + below,
			Baseline: supBox.Height + bb.Baseline,
		}
	case sup != nil:
		// Exponent sits to the upper-right of the base.
		return box{
			Width:    w,
			Height:   bb.Height + supBox.Height,
			Baseline: supBox.Height + bb.Baseline,
		}
	default: // sub only
		// The sub starts on the base's last row (or the row below a
		// single-row base) and may extend past the base's bottom.
		subTop := bb.Height
		if bb.Height > 1 {
			subTop = bb.Height - 1
		}
		return box{Width: w, Height: max(bb.Height, subTop+subBox.Height), Baseline: bb.Baseline}
	}
}

// Radicals render as √x for a bare symbol or number, √(…) when the
// radicand is compound, and √ with tall parens when it spans several
// rows. Keeping the radical on one row (rather than drawing a rule
// over the radicand) costs no vertical space and stays unambiguous.

// radicandBare reports whether the radicand can follow √ without
// parentheses. ASCII mode always uses parens because its radical
// glyph (a backslash) is not self-evidently a radical.
func radicandBare(n *node, s renderCtx) bool {
	if n == nil || s.ASCII {
		return false
	}
	return n.Type == nodeSymbol || n.Type == nodeNumber
}

// radicalWidth is the width of `√` plus the radicand and its parens.
func radicalWidth(inner box, content *node, s renderCtx) int {
	if inner.Height <= 1 && radicandBare(content, s) {
		return 1 + inner.Width
	}
	return 1 + 1 + inner.Width + 1
}

func measureSqrt(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	w := radicalWidth(inner, n.Children[0], s)
	if inner.Height <= 1 {
		return box{Width: w, Height: 1, Baseline: 0}
	}
	return box{Width: w, Height: inner.Height, Baseline: inner.Baseline}
}

func measureNthRoot(n *node, s renderCtx) box {
	nWidth := nthRootIndexWidth(n.Children[0], s)
	inner := measure(n.Children[1], s)
	w := nWidth + radicalWidth(inner, n.Children[1], s)

	// The index sits to the left of √, baseline-aligned with the
	// radical. If the index is multi-row (e.g. a fraction), it can
	// extend above and/or below the radicand's own rows and grows the
	// construct accordingly.
	var idx box
	if !canInlineSuperscript(n.Children[0], s) {
		idx = measure(n.Children[0], s)
	}

	innerTopRows := inner.Baseline
	innerBotRows := inner.Height - inner.Baseline - 1
	if inner.Height <= 1 {
		innerTopRows, innerBotRows = 0, 0
	}

	topRows := innerTopRows
	if idx.Baseline > topRows {
		topRows = idx.Baseline
	}
	idxBotRows := idx.Height - idx.Baseline - 1
	if idxBotRows < 0 {
		idxBotRows = 0
	}
	botRows := innerBotRows
	if idxBotRows > botRows {
		botRows = idxBotRows
	}

	h := topRows + 1 + botRows
	return box{Width: w, Height: h, Baseline: topRows}
}

// nthRootIndexWidth returns the rendered width of the nth-root index. When
// the index is inlineable (digits, simple symbols), it collapses to its
// Unicode superscript form (e.g. "3" → "³") which is 1 cell wide.
// The nth-root index always uses the Unicode superscript form when
// is no meaningful stacked alternative for a radical's index.
func nthRootIndexWidth(n *node, s renderCtx) int {
	if canInlineSuperscript(n, s) {
		return inlineScriptWidth(n)
	}
	return measure(n, s).Width
}

func measureParen(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	openW := displayWidth(s.displayValue(n.Open))
	closeW := displayWidth(s.displayValue(n.Close))
	if inner.Height <= 1 {
		// Simple case: single-line parens
		return box{
			Width:    openW + inner.Width + closeW,
			Height:   1,
			Baseline: 0,
		}
	}
	// Tall delimiters: a known delimiter takes one column; an unknown
	// one is stamped verbatim on every row and keeps its width.
	return box{
		Width:    openW + inner.Width + closeW,
		Height:   inner.Height,
		Baseline: inner.Baseline,
	}
}

// matrixColGap is the number of blank cells before column j. Aligned
// environments glue each right/left column pair together (`a &= b`)
// and separate pairs with a wider gap.
func matrixColGap(n *node, j int) int {
	if j == 0 {
		return 0
	}
	if n.Value == "align" {
		if j%2 == 1 {
			return 1
		}
		return 4
	}
	return 2
}

// matrixColAlign returns 'l', 'c' or 'r' for column j.
func matrixColAlign(n *node, j int) byte {
	switch n.Value {
	case "align":
		if j%2 == 0 {
			return 'r'
		}
		return 'l'
	case "cases", "lines":
		return 'l'
	case "array":
		if j < len(n.Cols) {
			return n.Cols[j]
		}
	}
	return 'c'
}

// matrixGrid measures every cell and returns per-column widths and
// per-row heights and baselines. Shared by measure and render.
func matrixGrid(n *node, s renderCtx) (colWidths, rowHeights, rowBaselines []int) {
	nrows := len(n.Rows)
	ncols := 0
	for _, row := range n.Rows {
		if len(row) > ncols {
			ncols = len(row)
		}
	}
	colWidths = make([]int, ncols)
	rowHeights = make([]int, nrows)
	rowBaselines = make([]int, nrows)
	for i, row := range n.Rows {
		// Cells are baseline-aligned within the row, so the row needs
		// the deepest ascent plus the deepest descent — not just the
		// tallest cell. (`0^0 & 0_0`: one cell rises, the other sinks.)
		above, below := 0, 0
		for j, cell := range row {
			b := measure(cell, s)
			if b.Width > colWidths[j] {
				colWidths[j] = b.Width
			}
			if b.Baseline > above {
				above = b.Baseline
			}
			if d := b.Height - b.Baseline - 1; d > below {
				below = d
			}
		}
		rowHeights[i] = above + 1 + below
		rowBaselines[i] = above
	}
	return colWidths, rowHeights, rowBaselines
}

func measureMatrix(n *node, s renderCtx) box {
	if len(n.Rows) == 0 {
		return box{Width: 2, Height: 1, Baseline: 0}
	}
	colWidths, rowHeights, _ := matrixGrid(n, s)

	totalW := 0
	for j, cw := range colWidths {
		totalW += matrixColGap(n, j) + cw
	}

	totalH := 0
	for _, rh := range rowHeights {
		totalH += rh
	}

	// Add delimiters. Each delimiter takes 1 cell plus 1 cell of padding
	// inside the matrix. Matches renderMatrix's column placement exactly.
	if n.Open != "" {
		totalW += 2
	}
	if n.Close != "" {
		totalW += 2
	}

	baseline := totalH / 2
	return box{Width: totalW, Height: totalH, Baseline: baseline}
}

func measureBigOp(n *node, s renderCtx) box {
	w := displayWidth(s.displayValue(n.Value))
	if w < 1 {
		w = 1
	}
	return box{Width: w, Height: 1, Baseline: 0}
}

func measureOverline(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	// An empty base (`\widehat{}`) still shows the mark.
	return box{Width: max(inner.Width, 1), Height: inner.Height + 1, Baseline: inner.Baseline + 1}
}

func measureUnderline(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	return box{Width: inner.Width, Height: inner.Height + 1, Baseline: inner.Baseline}
}

func measureHat(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	if canCombineAccent(n.Children[0], s) {
		// Combining marks are zero-width; the cell stays the same size.
		return inner
	}
	// An empty base (`\hat{}`) still shows the mark.
	return box{Width: max(inner.Width, 1), Height: inner.Height + 1, Baseline: inner.Baseline + 1}
}

func measureOverbrace(n *node, s renderCtx) box {
	expr := measure(n.Children[0], s)
	width := expr.Width
	height := 1 + expr.Height
	baseline := 1 + expr.Baseline
	if len(n.Children) >= 2 && n.Children[1] != nil {
		label := measure(n.Children[1], s)
		if label.Width > width {
			width = label.Width
		}
		height += label.Height
		baseline += label.Height
	}
	return box{Width: width, Height: height, Baseline: baseline}
}

func measureUnderbrace(n *node, s renderCtx) box {
	expr := measure(n.Children[0], s)
	width := expr.Width
	height := expr.Height + 1
	baseline := expr.Baseline
	if len(n.Children) >= 2 && n.Children[1] != nil {
		label := measure(n.Children[1], s)
		if label.Width > width {
			width = label.Width
		}
		height += label.Height
	}
	return box{Width: width, Height: height, Baseline: baseline}
}

func displayWidth(s string) int {
	w := 0
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		rw := runeWidth(r)
		if rw == 0 && w == 0 {
			// A combining mark with no base occupies a cell of its
			// own; putStrColored paints it the same way.
			rw = 1
		}
		w += rw
	}
	return w
}

func runeWidth(r rune) int {
	// Combining marks are painted on top of the preceding cell.
	if (r >= 0x0300 && r <= 0x036F) || (r >= 0x20D0 && r <= 0x20FF) {
		return 0
	}
	// Most mathematical symbols are single-width in modern terminals.
	// CJK and some special chars are double-width, but we handle the common case.
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2 // fullwidth forms
	}
	if r >= 0x4E00 && r <= 0x9FFF {
		return 2 // CJK
	}
	return 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
