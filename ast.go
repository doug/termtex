package termtex

// nodeType identifies the kind of AST node.
type nodeType int

const (
	nodeSymbol     nodeType = iota // single character or multi-char name
	nodeNumber                     // numeric literal
	nodeOperator                   // +, -, =, etc.
	nodeGroup                      // horizontal sequence of nodes
	nodeFrac                       // \frac{num}{den}
	nodeScript                     // base with optional sub/sup; Children=[base,sub,sup] (nil for absent)
	nodeSqrt                       // \sqrt{expr}
	nodeNthRoot                    // \sqrt[n]{expr}
	nodeParen                      // delimited group: (, ), [, ], {, }, |
	nodeMatrix                     // matrix/pmatrix/bmatrix environments
	nodeText                       // \text{...}
	nodeSpace                      // explicit spacing commands
	nodeBigOp                      // \sum, \prod, \int, \oint — Value carries the symbol
	nodeLim                        // \lim with subscript
	nodeOverline                   // \overline{expr}, \widehat{expr}, \widetilde{expr}
	nodeUnderline                  // \underline{expr}
	nodeHat                        // \hat / \dot / \ddot / \tilde / \vec
	nodeOverbrace                  // \overbrace{expr}^{label}
	nodeUnderbrace                 // \underbrace{expr}_{label}
)

// node represents a single element in the math AST.
type node struct {
	Type     nodeType
	Value    string  // symbol text, operator character, delimiter, etc.
	Children []*node // meaning depends on Type (see below)

	// For delimited groups
	Open  string // opening delimiter
	Close string // closing delimiter

	// For matrix
	Rows [][]*node

	// For nodeSpace: width in cells (0 is meaningful — see \!)
	Width int
}

// Helper constructors

func symNode(s string) *node {
	return &node{Type: nodeSymbol, Value: s}
}

func numNode(s string) *node {
	return &node{Type: nodeNumber, Value: s}
}

func opNode(s string) *node {
	return &node{Type: nodeOperator, Value: s}
}

func groupNode(children ...*node) *node {
	return &node{Type: nodeGroup, Children: children}
}

func fracNode(num, den *node) *node {
	return &node{Type: nodeFrac, Children: []*node{num, den}}
}

// scriptNode builds a nodeScript with Children = [base, sub, sup].
// Either sub or sup (but not both) may be nil.
func scriptNode(base, sub, sup *node) *node {
	return &node{Type: nodeScript, Children: []*node{base, sub, sup}}
}

// scriptParts splits a nodeScript into its (base, sub, sup); sub or
// sup is nil when that script is absent.
func scriptParts(n *node) (base, sub, sup *node) {
	return n.Children[0], n.Children[1], n.Children[2]
}

func sqrtNode(expr *node) *node {
	return &node{Type: nodeSqrt, Children: []*node{expr}}
}

func nthRootNode(n, expr *node) *node {
	return &node{Type: nodeNthRoot, Children: []*node{n, expr}}
}

func parenNode(open, close string, inner *node) *node {
	return &node{Type: nodeParen, Open: open, Close: close, Children: []*node{inner}}
}

func textNode(s string) *node {
	return &node{Type: nodeText, Value: s}
}

func spaceNode(width int) *node {
	return &node{Type: nodeSpace, Width: width}
}
