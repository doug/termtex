package termtex

// lineIndent is the indent of continuation rows produced by breakLines.
const lineIndent = 4

// breakLines splits a top-level horizontal list wider than width into
// several rows: first before relations (`=`, `≤`, `→`), and, for rows
// that are still too wide, before binary operators. Continuation rows
// are indented. Returns n unchanged when it fits, when width is zero,
// or when there is nothing to break at.
func breakLines(n *node, s renderCtx, width int) *node {
	if n == nil || width <= 0 || n.Type != nodeGroup {
		return n
	}
	if measure(n, s).Width <= width {
		return n
	}
	var rows [][]*node
	for i, row := range packRows(n.Children, s, width, atomRel) {
		avail := width
		if i > 0 {
			avail -= lineIndent
		}
		if measure(rowGroup(row, i > 0), s).Width > avail {
			rows = append(rows, packRows(row, s, avail, atomBin)...)
		} else {
			rows = append(rows, row)
		}
	}
	if len(rows) <= 1 {
		return n
	}
	m := &node{Type: nodeMatrix, Value: "lines"}
	for i, row := range rows {
		g := rowGroup(row, i > 0)
		if i > 0 {
			g.Children = append([]*node{spaceNode(lineIndent)}, g.Children...)
		}
		m.Rows = append(m.Rows, []*node{g})
	}
	return m
}

// rowGroup wraps a row's children in a group. A continuation row that
// opens with a binary operator keeps the operator's trailing space
// (`+ e`), which the unary-demotion rule would otherwise drop.
func rowGroup(row []*node, cont bool) *node {
	children := append([]*node{}, row...)
	if cont && len(children) > 0 && leftClass(children[0]) == atomBin {
		children = append([]*node{children[0], spaceNode(1)}, children[1:]...)
	}
	return groupNode(children...)
}

// packRows cuts children before every atom of class cls and packs the
// pieces greedily into rows no wider than width (continuation rows
// lose lineIndent cells). A piece that is wider than a row on its own
// still gets its own row.
func packRows(children []*node, s renderCtx, width int, cls atomClass) [][]*node {
	lc, _, _ := effectiveClasses(children)
	var pieces [][]*node
	start := 0
	for i := 1; i < len(children); i++ {
		if lc[i] == cls {
			pieces = append(pieces, children[start:i])
			start = i
		}
	}
	pieces = append(pieces, children[start:])
	if len(pieces) == 1 {
		return [][]*node{children}
	}

	var rows [][]*node
	var cur []*node
	for _, piece := range pieces {
		avail := width
		if len(rows) > 0 {
			avail -= lineIndent
		}
		if len(cur) > 0 {
			cand := append(append([]*node{}, cur...), piece...)
			if measure(rowGroup(cand, len(rows) > 0), s).Width > avail {
				rows = append(rows, cur)
				cur = nil
			}
		}
		cur = append(cur, piece...)
	}
	if len(cur) > 0 {
		rows = append(rows, cur)
	}
	return rows
}
