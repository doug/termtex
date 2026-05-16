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
// the context's cache so repeated walks (initial measure + render) do
// not recompute. Cache keys are (node pointer, compact flag).
func measure(n *node, s renderCtx) box {
	if n == nil {
		return box{Width: 0, Height: 1, Baseline: 0}
	}
	if s.cache != nil {
		if b, ok := s.cache.get(n, s.compact); ok {
			return b
		}
	}
	b := measureUncached(n, s)
	if s.cache != nil {
		s.cache.put(n, s.compact, b)
	}
	return b
}

// measureUncached is the bottom-up measurement function. Callers
// should go through [measure] so results are memoized.
func measureUncached(n *node, s renderCtx) box {
	switch n.Type {
	case nodeSymbol, nodeNumber, nodeOperator, nodeText:
		return box{Width: displayWidth(s.displayValue(n.Value)), Height: 1, Baseline: 0}
	case nodeSpace:
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
		return box{Width: 4, Height: 1, Baseline: 0} // "lim" + trailing space
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

// isTextOp returns true if the node is a text operator (sin, cos, log, etc.)
// or has a text operator as its base (e.g., \sin^2).
func isTextOp(n *node) bool {
	if n.Type == nodeText {
		return true
	}
	if n.Type == nodeLim {
		return true
	}
	if n.Type == nodeScript && len(n.Children) > 0 {
		return n.Children[0].Type == nodeText || n.Children[0].Type == nodeLim
	}
	return false
}

// needsTextSpaceBefore returns true if child[i] needs a leading space due to text operator adjacency.
func needsTextSpaceBefore(children []*node, i int) bool {
	if i == 0 {
		return false
	}
	prev := children[i-1]
	cur := children[i]
	// Space before a text op when preceded by a symbol, number, paren, sub/sup, or another text op
	if isTextOp(cur) {
		switch prev.Type {
		case nodeSymbol, nodeNumber, nodeParen, nodeScript,
			nodeSqrt, nodeNthRoot, nodeFrac, nodeMatrix:
			return true
		}
		if isTextOp(prev) {
			return true
		}
	}
	// Space after a text op when followed by a symbol, number, paren, or multi-line construct
	if isTextOp(prev) {
		switch cur.Type {
		case nodeSymbol, nodeNumber, nodeParen, nodeSqrt, nodeNthRoot,
			nodeFrac, nodeMatrix, nodeScript:
			return true
		}
	}
	return false
}

// needsBarSeparator returns true when prev and cur would render with
// adjacent horizontal strokes that visually merge — e.g. a leading
// `-` operator before a fraction draws as `-────`, reading as one
// continuous line.
func needsBarSeparator(prev, cur *node) bool {
	if prev == nil || cur == nil || prev.Type != nodeOperator {
		return false
	}
	switch prev.Value {
	case "-", "+", "−", "±", "∓", "+/-", "-/+":
	default:
		return false
	}
	return cur.Type == nodeFrac
}

// hasTrailingGroupSpace reports whether child[i-1] already added a
// trailing space in the renderGroup loop. Used to avoid double-spacing
// when inserting a collision-guard space.
func hasTrailingGroupSpace(children []*node, i int, spaced bool) bool {
	if i < 1 {
		return false
	}
	pi := i - 1
	prev := children[pi]
	return spaced && isSpacedOp(prev) && pi > 0 && !isSpacedOp(children[pi-1])
}

// isSpacedOp returns true for operators that should have spaces around them in a group.
func isSpacedOp(n *node) bool {
	if n.Type != nodeOperator {
		return false
	}
	switch n.Value {
	case "=", "+", "-", "±", "∓", "×", "÷",
		"<", ">", "≤", "≥", "≠", "≈", "≡",
		"∈", "∉", "⊂", "⊆", "∪", "∩",
		"→", "←", "⇒", "⇐", "↔", "⇔",
		// ASCII-mode equivalents
		"+/-", "-/+", "<=", ">=", "!=", "~=", "==",
		"->", "<-", "=>", "<->", "<=>",
		"in", "!in", "sub", "sube", "sup", "supe", "U":
		return true
	}
	return false
}

// groupNeedsSpacing returns true if a group should add spaces around
// operators. In compact contexts (subscripts, superscripts, big-op
// limits) the rendering is tight (`i=1` not `i = 1`); the renderer
// signals this by setting ctx.compact, threaded down by the script
// rendering paths.
func groupNeedsSpacing(_ *node, s renderCtx) bool {
	return !s.compact
}

func measureGroup(n *node, s renderCtx) box {
	if len(n.Children) == 0 {
		// An empty group contributes no rows. Without this, constructs
		// that add label/index height (\overbrace, \sum^{}, \hat{}, …)
		// reserve a phantom blank row in the output.
		return box{Width: 0, Height: 0, Baseline: 0}
	}
	spaced := groupNeedsSpacing(n, s)
	totalWidth := 0
	maxAbove := 0 // rows above baseline
	maxBelow := 0 // rows below baseline (including baseline row)
	for i, child := range n.Children {
		b := measure(child, s)
		if spaced && isSpacedOp(child) && i > 0 && !isSpacedOp(n.Children[i-1]) {
			totalWidth += 2 // 1 space on each side
		}
		if needsTextSpaceBefore(n.Children, i) {
			totalWidth++ // text-operator spacing always applies
		}
		if i > 0 && !hasTrailingGroupSpace(n.Children, i, spaced) &&
			needsBarSeparator(n.Children[i-1], child) {
			totalWidth++ // collision guard: `-` adjacent to frac bar
		}
		totalWidth += b.Width
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

func measureFrac(n *node, s renderCtx) box {
	num := measure(n.Children[0], s)
	den := measure(n.Children[1], s)
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

// measureScript handles base[_sub][^sup]. Either sub or sup may be
// nil, and there are three distinct shapes depending on the base:
//
//   - big operators (\sum, \int, \lim): the limits stack above/below
//     the operator. A +1 trailing pad is added so subsequent content
//     doesn't collide with the limits.
//   - inline scripts (single-cell Unicode super/sub forms): rendered
//     to the right of the base on the same row.
//   - stacked scripts: super goes upper-right, sub goes lower-right.
func measureScript(n *node, s renderCtx) box {
	base, sub, sup := scriptParts(n)
	bb := measure(base, s)
	cs := s.withCompact()

	if isBigOp(base) {
		var subBox, supBox box
		if sub != nil {
			subBox = measure(sub, cs)
		}
		if sup != nil {
			supBox = measure(sup, cs)
		}
		// +1 trailing pad so limits don't collide with next element.
		w := max(bb.Width, max(subBox.Width, supBox.Width)) + 1
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
		return box{
			Width:    w,
			Height:   supBox.Height + bb.Height + subBox.Height - overlap,
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
		h := bb.Height + subBox.Height - 1
		if h < bb.Height {
			h = bb.Height
		}
		if subBox.Height >= bb.Height {
			h = subBox.Height + 1
		}
		return box{Width: w, Height: h, Baseline: bb.Baseline}
	}
}

// Sqrt is rendered as a function-like construct: √(content), with tall
// parens for multi-line content. This keeps a single rendering path so
// nested sqrt + frac + sup compose predictably with everything else.

func measureSqrt(n *node, s renderCtx) box {
	inner := measure(n.Children[0], s)
	// √ + ( + content + )
	w := 1 + 1 + inner.Width + 1
	if inner.Height <= 1 {
		return box{Width: w, Height: 1, Baseline: 0}
	}
	return box{Width: w, Height: inner.Height, Baseline: inner.Baseline}
}

func measureNthRoot(n *node, s renderCtx) box {
	nWidth := nthRootIndexWidth(n.Children[0], s)
	inner := measure(n.Children[1], s)
	w := nWidth + 1 + 1 + inner.Width + 1

	// The index sits to the left of √, baseline-aligned with the
	// radical. If the index is multi-row (e.g. a fraction), it can
	// extend above and/or below the radicand's own rows and grows the
	// construct accordingly.
	var idx box
	if !canInlineSupRaw(n.Children[0], s) {
		idx = measure(n.Children[0], s)
	}

	innerBaseline := inner.Baseline
	innerTopRows := inner.Baseline
	innerBotRows := inner.Height - inner.Baseline - 1
	if inner.Height <= 1 {
		innerTopRows, innerBotRows, innerBaseline = 0, 0, 0
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
	_ = innerBaseline
	return box{Width: w, Height: h, Baseline: topRows}
}

// nthRootIndexWidth returns the rendered width of the nth-root index. When
// the index is inlineable (digits, simple symbols), it collapses to its
// Unicode superscript form (e.g. "3" → "³") which is 1 cell wide.
// The nth-root index always uses the Unicode superscript form when
// possible — it doesn't participate in forceStackScripts because there
// is no meaningful stacked alternative for a radical's index.
func nthRootIndexWidth(n *node, s renderCtx) int {
	if canInlineSupRaw(n, s) {
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
	// Tall delimiters: each delimiter takes 1 column
	dw := 0
	if n.Open != "" {
		dw++
	}
	if n.Close != "" {
		dw++
	}
	return box{
		Width:    dw + inner.Width,
		Height:   inner.Height,
		Baseline: inner.Baseline,
	}
}

func measureMatrix(n *node, s renderCtx) box {
	if len(n.Rows) == 0 {
		return box{Width: 2, Height: 1, Baseline: 0}
	}
	nrows := len(n.Rows)
	ncols := 0
	for _, row := range n.Rows {
		if len(row) > ncols {
			ncols = len(row)
		}
	}

	// Measure each cell, find max width per column and max height per row
	colWidths := make([]int, ncols)
	rowHeights := make([]int, nrows)
	rowBaselines := make([]int, nrows)
	for i, row := range n.Rows {
		rowHeights[i] = 1
		for j, cell := range row {
			b := measure(cell, s)
			if b.Width > colWidths[j] {
				colWidths[j] = b.Width
			}
			if b.Height > rowHeights[i] {
				rowHeights[i] = b.Height
			}
			if b.Baseline > rowBaselines[i] {
				rowBaselines[i] = b.Baseline
			}
		}
	}

	totalW := 0
	for _, cw := range colWidths {
		totalW += cw + 2 // 2-char gap between columns
	}
	if totalW > 0 {
		totalW -= 2 // no trailing gap
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
	return box{Width: inner.Width, Height: inner.Height + 1, Baseline: inner.Baseline + 1}
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
	return box{Width: inner.Width, Height: inner.Height + 1, Baseline: inner.Baseline + 1}
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
		w += runeWidth(r)
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
