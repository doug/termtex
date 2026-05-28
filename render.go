package termtex

import (
	"strings"
	"sync"
	"unicode/utf8"
)

// cell holds a rune, an optional combining mark (for accents like
// \hat, \tilde rendered on top of the base glyph), and an optional
// ANSI color prefix.
type cell struct {
	ch     rune
	accent rune   // combining mark output after ch; 0 for none
	color  string // ANSI escape, empty for no color
}

// canvas is a 2D character buffer for compositing output. Backed by a
// contiguous []cell with row stride = width — one allocation per
// canvas, recycled through canvasPool.
type canvas struct {
	cells  []cell
	width  int
	height int
	ctx    renderCtx
}

var canvasPool = sync.Pool{
	New: func() any { return &canvas{} },
}

// spaceCells is a long buffer of blank cells used as the source for
// copy()-based canvas clearing. Sized to cover the largest expected
// canvas; copy() trims to whatever target we hand it.
var spaceCells = func() []cell {
	s := make([]cell, 4096)
	blank := cell{ch: ' '}
	for i := range s {
		s[i] = blank
	}
	return s
}()

func newCanvas(w, h int) *canvas {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	c := canvasPool.Get().(*canvas)
	size := w * h
	if cap(c.cells) < size {
		c.cells = make([]cell, size)
	} else {
		c.cells = c.cells[:size]
	}
	// Fast clear: copy from a pre-blanked buffer in chunks. cheaper
	// than a per-cell assignment loop because copy() turns into a
	// runtime memmove.
	for off := 0; off < size; {
		n := copy(c.cells[off:], spaceCells)
		off += n
	}
	c.width = w
	c.height = h
	c.ctx = renderCtx{}
	return c
}

// release returns the canvas to the pool. The caller must not
// reference its cells slice after calling.
func (c *canvas) release() {
	canvasPool.Put(c)
}

func (c *canvas) set(x, y int, r rune) {
	c.setColored(x, y, r, "")
}

func (c *canvas) setColored(x, y int, r rune, color string) {
	if y >= 0 && y < c.height && x >= 0 && x < c.width {
		c.cells[y*c.width+x] = cell{ch: r, color: color}
	}
}

// withCompactCtx runs fn with the canvas's compact flag set, restoring
// it after. Used when rendering script content (sub/sup, big-op limits)
// where operator spacing should be suppressed.
func (c *canvas) withCompactCtx(fn func()) {
	saved := c.ctx.compact
	c.ctx.compact = true
	fn()
	c.ctx.compact = saved
}

// setAccent attaches a combining mark (e.g. U+0302 for hat) to the
// cell at (x, y). The mark is emitted right after the base rune in
// String() so terminals render it on top of the existing character.
func (c *canvas) setAccent(x, y int, mark rune) {
	if y >= 0 && y < c.height && x >= 0 && x < c.width {
		c.cells[y*c.width+x].accent = mark
	}
}

func (c *canvas) putStr(x, y int, s string) {
	c.putStrColored(x, y, s, "")
}

func (c *canvas) putStrColored(x, y int, s string, color string) {
	col := x
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		c.setColored(col, y, r, color)
		col += runeWidth(r)
	}
}

func (c *canvas) hline(x, y, w int, ch rune) {
	c.hlineColored(x, y, w, ch, "")
}

func (c *canvas) hlineColored(x, y, w int, ch rune, color string) {
	for i := 0; i < w; i++ {
		c.setColored(x+i, y, ch, color)
	}
}

func (c *canvas) vline(x, y, h int, ch rune) {
	c.vlineColored(x, y, h, ch, "")
}

func (c *canvas) vlineColored(x, y, h int, ch rune, color string) {
	for i := 0; i < h; i++ {
		c.setColored(x, y+i, ch, color)
	}
}

// String renders the canvas to a string, trimming trailing whitespace per line.
func (c *canvas) String() string {
	var sb strings.Builder
	// Heuristic preallocation: most cells are 1-3 UTF-8 bytes; aim for
	// width*height with a small overhead, plus newlines.
	sb.Grow(c.width*c.height + c.height)
	useColor := c.ctx.Color
	w := c.width
	for i := 0; i < c.height; i++ {
		if i > 0 {
			sb.WriteByte('\n')
		}
		row := c.cells[i*w : (i+1)*w]
		// Find last non-space
		last := len(row) - 1
		for last >= 0 && row[last].ch == ' ' {
			last--
		}
		prevColor := ""
		for j := 0; j <= last; j++ {
			cl := row[j]
			if useColor {
				if cl.color != prevColor {
					if cl.color == "" {
						sb.WriteString(ansiReset)
					} else {
						sb.WriteString(cl.color)
					}
					prevColor = cl.color
				}
			}
			sb.WriteRune(cl.ch)
			if cl.accent != 0 {
				sb.WriteRune(cl.accent)
			}
		}
		if useColor && prevColor != "" {
			sb.WriteString(ansiReset)
			prevColor = ""
		}
	}
	return sb.String()
}

// renderNode paints an AST node onto a canvas at position (x, y).
func renderNode(c *canvas, n *node, x, y int) {
	if n == nil {
		return
	}
	switch n.Type {
	case nodeSymbol:
		s := c.ctx.displayValue(n.Value)
		if c.ctx.Italic && !c.ctx.ASCII {
			s = mathItalic(s)
		}
		c.putStrColored(x, y, s, c.ctx.color(semVariable))
	case nodeNumber:
		c.putStrColored(x, y, c.ctx.displayValue(n.Value), c.ctx.color(semNumber))
	case nodeOperator:
		c.putStrColored(x, y, c.ctx.displayValue(n.Value), c.ctx.color(semOperator))
	case nodeText:
		c.putStrColored(x, y, c.ctx.displayValue(n.Value), c.ctx.color(semText))
	case nodeSpace:
		// just skip width cells
	case nodeGroup:
		renderGroup(c, n, x, y)
	case nodeFrac:
		renderFrac(c, n, x, y)
	case nodeScript:
		renderScript(c, n, x, y)
	case nodeSqrt:
		renderSqrt(c, n, x, y)
	case nodeNthRoot:
		renderNthRoot(c, n, x, y)
	case nodeParen:
		renderParen(c, n, x, y)
	case nodeMatrix:
		renderMatrix(c, n, x, y)
	case nodeBigOp:
		c.putStrColored(x, y, c.ctx.displayValue(n.Value), c.ctx.color(semBigOp))
	case nodeLim:
		c.putStrColored(x, y, "lim", c.ctx.color(semText))
	case nodeOverline:
		renderOverline(c, n, x, y)
	case nodeUnderline:
		renderUnderline(c, n, x, y)
	case nodeHat:
		renderHat(c, n, x, y)
	case nodeOverbrace:
		renderOverbrace(c, n, x, y)
	case nodeUnderbrace:
		renderUnderbrace(c, n, x, y)
	}
}

func renderGroup(c *canvas, n *node, x, y int) {
	box := measure(n, c.ctx)
	// Binary-operator spacing (`a + b`) is suppressed in compact
	// contexts (subscripts, big-op limits) so `i=1` stays tight.
	// Text-operator spacing (`sin x`) stays on regardless — words need
	// to remain readable.
	spaced := groupNeedsSpacing(n, c.ctx)
	cx := x
	for i, child := range n.Children {
		cb := measure(child, c.ctx)
		if spaced && isSpacedOp(child) && i > 0 && !isSpacedOp(n.Children[i-1]) {
			cx++
		}
		if needsTextSpaceBefore(n.Children, i) {
			cx++
		}
		if i > 0 && !hasTrailingGroupSpace(n.Children, i, spaced) &&
			needsBarSeparator(n.Children[i-1], child) {
			cx++ // collision guard: `-` adjacent to frac bar
		}
		dy := box.Baseline - cb.Baseline
		renderNode(c, child, cx, y+dy)
		cx += cb.Width
		if spaced && isSpacedOp(child) && i > 0 && !isSpacedOp(n.Children[i-1]) {
			cx++
		}
	}
}

func renderFrac(c *canvas, n *node, x, y int) {
	box := measure(n, c.ctx)
	num := measure(n.Children[0], c.ctx)
	den := measure(n.Children[1], c.ctx)

	barY := y + num.Height
	barW := box.Width

	// Center numerator above bar
	numX := x + (barW-num.Width)/2
	renderNode(c, n.Children[0], numX, y)

	// Draw fraction bar
	c.hlineColored(x, barY, barW, c.ctx.glyphs().FracBar, c.ctx.color(semBar))

	// Center denominator below bar
	denX := x + (barW-den.Width)/2
	renderNode(c, n.Children[1], denX, barY+1)
}

// renderScript paints base[_sub][^sup]. Mirrors measureScript's three
// shapes: big-operator stack, inline single-cell scripts, and stacked
// upper-right/lower-right scripts.
func renderScript(c *canvas, n *node, x, y int) {
	base, sub, sup := scriptParts(n)
	bb := measure(base, c.ctx)
	cs := c.ctx.withCompact()

	if isBigOp(base) {
		var subBox, supBox box
		if sub != nil {
			subBox = measure(sub, cs)
		}
		if sup != nil {
			supBox = measure(sup, cs)
		}
		contentW := max(bb.Width, max(subBox.Width, supBox.Width))
		c.withCompactCtx(func() {
			if sup != nil {
				renderNode(c, sup, x+(contentW-supBox.Width)/2, y)
			}
			if sub != nil {
				renderNode(c, sub, x+(contentW-subBox.Width)/2, y+supBox.Height+bb.Height)
			}
		})
		renderNode(c, base, x+(contentW-bb.Width)/2, y+supBox.Height)
		return
	}

	// The inline shortcut only works when the base is a single row; a
	// multi-row base (e.g. a fraction) needs the stacked path so the
	// script lands at the base's baseline row rather than its top.
	supInline := bb.Height == 1 && sup != nil && canInlineSuperscript(sup, c.ctx)
	subInline := bb.Height == 1 && sub != nil && canInlineSubscript(sub, c.ctx)

	if (sup != nil && sub == nil && supInline) ||
		(sub != nil && sup == nil && subInline) ||
		(sup != nil && sub != nil && supInline && subInline) {
		renderNode(c, base, x, y)
		cx := x + bb.Width
		if sup != nil {
			s := toSuperscript(sup)
			c.putStrColored(cx, y, s, c.ctx.color(semNumber))
			cx += displayWidth(s)
		}
		if sub != nil {
			c.putStrColored(cx, y, toSubscript(sub), c.ctx.color(semVariable))
		}
		return
	}

	// Stacked. Place the base on the script-aware baseline row.
	box := measure(n, c.ctx)
	baseY := y + box.Baseline - bb.Baseline
	renderNode(c, base, x, baseY)

	c.withCompactCtx(func() {
		if sup != nil {
			supBox := measure(sup, cs)
			supY := baseY - supBox.Height
			if sub == nil {
				// Sup-only: place at top of construct.
				supY = y
			}
			if supY < y {
				supY = y
			}
			renderNode(c, sup, x+bb.Width, supY)
		}
		if sub != nil {
			subY := baseY + bb.Height
			if bb.Height > 1 {
				subY = baseY + bb.Height - 1
			}
			renderNode(c, sub, x+bb.Width, subY)
		}
	})
}

func renderSqrt(c *canvas, n *node, x, y int) {
	renderRadical(c, x, y, 0, n.Children[0])
}

func renderNthRoot(c *canvas, n *node, x, y int) {
	nWidth := nthRootIndexWidth(n.Children[0], c.ctx)
	cBox := measure(n, c.ctx)
	inner := measure(n.Children[1], c.ctx)

	// Radical baseline row within the construct, and the row where the
	// radicand's top lands so its own baseline aligns.
	baseY := y + cBox.Baseline
	innerY := baseY - inner.Baseline
	if inner.Height <= 1 {
		innerY = baseY
	}

	if canInlineSupRaw(n.Children[0], c.ctx) {
		c.putStrColored(x, baseY, toSuperscript(n.Children[0]), c.ctx.color(semNumber))
	} else {
		idx := measure(n.Children[0], c.ctx)
		renderNode(c, n.Children[0], x, baseY-idx.Baseline)
	}

	renderRadical(c, x+nWidth, innerY, 0, n.Children[1])
}

// renderRadical paints `√(content)` at (x, y), using tall parens when
// content has multiple rows. The √ sits on the content's baseline row.
func renderRadical(c *canvas, x, y, _ int, content *node) {
	inner := measure(content, c.ctx)
	dc := c.ctx.color(semDelim)
	g := c.ctx.glyphs()

	if inner.Height <= 1 {
		c.setColored(x, y, g.Sqrt, dc)
		c.setColored(x+1, y, '(', dc)
		renderNode(c, content, x+2, y)
		c.setColored(x+2+inner.Width, y, ')', dc)
		return
	}
	baseY := y + inner.Baseline
	c.setColored(x, baseY, g.Sqrt, dc)
	renderTallDelim(c, x+1, y, inner.Height, "(", dc)
	renderNode(c, content, x+2, y)
	renderTallDelim(c, x+2+inner.Width, y, inner.Height, ")", dc)
}

func renderParen(c *canvas, n *node, x, y int) {
	inner := measure(n.Children[0], c.ctx)
	open := c.ctx.displayValue(n.Open)
	close := c.ctx.displayValue(n.Close)
	openW := displayWidth(open)
	delimColor := c.ctx.color(semDelim)

	if inner.Height <= 1 {
		if open != "" {
			c.putStrColored(x, y, open, delimColor)
		}
		renderNode(c, n.Children[0], x+openW, y)
		if close != "" {
			c.putStrColored(x+openW+inner.Width, y, close, delimColor)
		}
		return
	}

	cx := x
	if open != "" {
		renderTallDelim(c, cx, y, inner.Height, open, delimColor)
		cx++
	}
	renderNode(c, n.Children[0], cx, y)
	cx += inner.Width
	if close != "" {
		renderTallDelim(c, cx, y, inner.Height, close, delimColor)
	}
}

// tallParts is the (top, middle, bottom) glyph triple for a simple
// tall delimiter — parens and brackets share this shape, where the
// middle glyph fills every row between top and bottom.
type tallParts struct{ T, M, B rune }

// pickTallParts resolves delim to its glyph triple. Returns ok=false
// for braces (special middle marker) and for vertical bars (uniform
// fill) — those callers handle the shape themselves.
func pickTallParts(delim string, g glyphs) (tallParts, bool) {
	switch delim {
	case "(":
		return tallParts{g.ParenLT, g.ParenLM, g.ParenLB}, true
	case ")":
		return tallParts{g.ParenRT, g.ParenRM, g.ParenRB}, true
	case "[":
		return tallParts{g.BrackLT, g.BrackLM, g.BrackLB}, true
	case "]":
		return tallParts{g.BrackRT, g.BrackRM, g.BrackRB}, true
	}
	return tallParts{}, false
}

func renderTallDelim(c *canvas, x, y, h int, delim string, color string) {
	g := c.ctx.glyphs()
	if h <= 1 {
		c.putStrColored(x, y, delim, color)
		return
	}
	// Uniform-fill delimiters (single and double vertical bars).
	switch delim {
	case "|":
		c.vlineColored(x, y, h, g.VBar, color)
		return
	case "‖":
		c.vlineColored(x, y, h, g.VBarDouble, color)
		return
	case "{":
		renderTallBrace(c, x, y, h, color, g.BraceLT, g.BraceLE, g.BraceLM, g.BraceLB)
		return
	case "}":
		renderTallBrace(c, x, y, h, color, g.BraceRT, g.BraceRE, g.BraceRM, g.BraceRB)
		return
	}
	parts, ok := pickTallParts(delim, g)
	if !ok {
		// Unknown delimiter: stamp the literal string on every row.
		for i := 0; i < h; i++ {
			c.putStrColored(x, y+i, delim, color)
		}
		return
	}
	c.setColored(x, y, parts.T, color)
	for i := 1; i < h-1; i++ {
		c.setColored(x, y+i, parts.M, color)
	}
	c.setColored(x, y+h-1, parts.B, color)
}

// renderTallBrace paints a brace with a centered connector glyph
// (BraceLM/BraceRM) and edge-fill glyphs (BraceLE/BraceRE) above and
// below. Height 2 collapses to top+bottom — no middle extension.
func renderTallBrace(c *canvas, x, y, h int, color string, top, edge, mid, bot rune) {
	c.setColored(x, y, top, color)
	c.setColored(x, y+h-1, bot, color)
	if h == 2 {
		return
	}
	m := h / 2
	for i := 1; i < h-1; i++ {
		ch := edge
		if i == m {
			ch = mid
		}
		c.setColored(x, y+i, ch, color)
	}
}

func renderMatrix(c *canvas, n *node, x, y int) {
	delimColor := c.ctx.color(semDelim)
	open := c.ctx.displayValue(n.Open)
	close := c.ctx.displayValue(n.Close)
	if len(n.Rows) == 0 {
		// measureMatrix reserves 2 cells for an empty matrix; paint the
		// delimiters so they don't silently disappear.
		if open != "" {
			c.putStrColored(x, y, open, delimColor)
		}
		if close != "" {
			c.putStrColored(x+displayWidth(open), y, close, delimColor)
		}
		return
	}
	nrows := len(n.Rows)
	ncols := 0
	for _, row := range n.Rows {
		if len(row) > ncols {
			ncols = len(row)
		}
	}

	colWidths := make([]int, ncols)
	rowHeights := make([]int, nrows)
	rowBaselines := make([]int, nrows)
	for i, row := range n.Rows {
		rowHeights[i] = 1
		for j, cell := range row {
			b := measure(cell, c.ctx)
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

	box := measure(n, c.ctx)
	cx := x
	if open != "" {
		renderTallDelim(c, cx, y, box.Height, open, delimColor)
		cx += 2 // 1 for the delimiter + 1 for the inside padding
	}

	ry := y
	for i, row := range n.Rows {
		colX := cx
		for j, cl := range row {
			cb := measure(cl, c.ctx)
			pad := (colWidths[j] - cb.Width) / 2
			// Baseline-align each cell within its row so that a fraction
			// and a plain symbol in the same row line up at the fraction
			// bar rather than at the row's top.
			cellY := ry + rowBaselines[i] - cb.Baseline
			renderNode(c, cl, colX+pad, cellY)
			colX += colWidths[j] + 2
		}
		ry += rowHeights[i]
	}

	if close != "" {
		// Last cell ended at cx + sum(cw+2) - 2 (no trailing gap after).
		// Leave a 1-cell pad before the closing delim — matches the
		// 1-cell pad after the opening delim and matches measureMatrix.
		closeX := cx - 2
		for _, cw := range colWidths {
			closeX += cw + 2
		}
		renderTallDelim(c, closeX+1, y, box.Height, close, delimColor)
	}
}

// accentMark resolves a nodeHat kind tag to a glyph rune.
func accentMark(g glyphs, kind string) rune {
	switch kind {
	case "dot":
		return g.DotMark
	case "ddot":
		return g.DDotMark
	case "tilde":
		return g.TildeMark
	case "vec":
		return g.VecMark
	default:
		return g.HatMark
	}
}

// combiningAccentMark returns the combining (zero-width) Unicode mark
// for an accent. These are placed after a base character so terminals
// render the mark on top of the glyph in the same cell.
func combiningAccentMark(kind string) rune {
	switch kind {
	case "dot":
		return '\u0307' // ̇
	case "ddot":
		return '\u0308' // ̈
	case "tilde":
		return '\u0303' // ̃
	case "vec":
		return '\u20D7' // ⃗
	default:
		return '\u0302' // ̂
	}
}

// canCombineAccent reports whether an accent's base content is a
// single-cell glyph that can carry a combining mark (so we can render
// accents in one row instead of stacking the mark on a row above).
func canCombineAccent(n *node, s renderCtx) bool {
	if n == nil || s.ASCII {
		return false
	}
	switch n.Type {
	case nodeSymbol, nodeNumber:
		return displayWidth(n.Value) == 1
	case nodeGroup:
		if len(n.Children) == 1 {
			return canCombineAccent(n.Children[0], s)
		}
	}
	return false
}

// wideAccentMark resolves a nodeOverline kind tag to a glyph rune.
func wideAccentMark(g glyphs, kind string) rune {
	switch kind {
	case "hat":
		return g.HatMark
	case "tilde":
		return g.TildeMark
	default:
		return g.OverBar
	}
}

func renderOverline(c *canvas, n *node, x, y int) {
	inner := measure(n.Children[0], c.ctx)
	c.hlineColored(x, y, inner.Width, wideAccentMark(c.ctx.glyphs(), n.Value), c.ctx.color(semBar))
	renderNode(c, n.Children[0], x, y+1)
}

func renderUnderline(c *canvas, n *node, x, y int) {
	inner := measure(n.Children[0], c.ctx)
	renderNode(c, n.Children[0], x, y)
	c.hlineColored(x, y+inner.Height, inner.Width, c.ctx.glyphs().UnderBar, c.ctx.color(semBar))
}

func renderHat(c *canvas, n *node, x, y int) {
	if canCombineAccent(n.Children[0], c.ctx) {
		renderNode(c, n.Children[0], x, y)
		c.setAccent(x, y, combiningAccentMark(n.Value))
		return
	}
	inner := measure(n.Children[0], c.ctx)
	hatX := x + inner.Width/2
	c.set(hatX, y, accentMark(c.ctx.glyphs(), n.Value))
	renderNode(c, n.Children[0], x, y+1)
}

func renderOverbrace(c *canvas, n *node, x, y int)  { renderBrace(c, n, x, y, true) }
func renderUnderbrace(c *canvas, n *node, x, y int) { renderBrace(c, n, x, y, false) }

// renderBrace paints \overbrace or \underbrace. The label (if any)
// sits above the brace when over=true, below when over=false; the
// expression goes on the opposite side.
func renderBrace(c *canvas, n *node, x, y int, over bool) {
	expr := measure(n.Children[0], c.ctx)
	construct := measure(n, c.ctx)
	barColor := c.ctx.color(semBar)
	hasLabel := len(n.Children) >= 2 && n.Children[1] != nil

	braceX := x + (construct.Width-expr.Width)/2
	exprX := braceX
	var braceY, exprY, labelY int
	if over {
		var labelHeight int
		if hasLabel {
			labelHeight = measure(n.Children[1], c.ctx).Height
		}
		braceY = y + labelHeight
		exprY = braceY + 1
		labelY = y
	} else {
		exprY = y
		braceY = y + expr.Height
		labelY = braceY + 1
	}

	if hasLabel {
		label := measure(n.Children[1], c.ctx)
		labelX := x + (construct.Width-label.Width)/2
		renderNode(c, n.Children[1], labelX, labelY)
	}
	drawBraceRow(c, braceX, braceY, expr.Width, over, c.ctx.glyphs(), barColor)
	renderNode(c, n.Children[0], exprX, exprY)
}

// drawBraceRow paints a single-row brace decoration: corners, fill, and
// a centered connector mark. `over` selects the overbrace glyph set
// (curving down toward the expr below) versus underbrace (curving up).
func drawBraceRow(c *canvas, x, y, width int, over bool, g glyphs, color string) {
	if width <= 0 {
		return
	}
	left := g.OverbraceLeft
	mid := g.OverbraceMid
	right := g.OverbraceRight
	if !over {
		left = g.UnderbraceLeft
		mid = g.UnderbraceMid
		right = g.UnderbraceRight
	}
	if width == 1 {
		c.setColored(x, y, mid, color)
		return
	}
	c.hlineColored(x, y, width, g.FracBar, color)
	c.setColored(x, y, left, color)
	c.setColored(x+width-1, y, right, color)
	if width >= 3 {
		c.setColored(x+width/2, y, mid, color)
	}
}
