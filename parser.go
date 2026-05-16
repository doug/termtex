package termtex

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType classifies lexer tokens.
type tokenType int

const (
	tokEOF tokenType = iota
	tokSymbol
	tokNumber
	tokOperator
	tokCommand    // \frac, \sqrt, etc.
	tokLBrace     // {
	tokRBrace     // }
	tokLBracket   // [
	tokRBracket   // ]
	tokLParen     // (
	tokRParen     // )
	tokPipe       // |
	tokCaret      // ^
	tokUnderscore // _
	tokAmpersand  // &
	tokNewline    // \\
	tokSpace
	tokText // raw text-mode content (inside \text{...})
)

// spaceWidths is the canonical width table for LaTeX spacing commands.
// The lexer emits these as tokSpace; the parser reads the width here.
var spaceWidths = map[string]int{
	"\\,":     1,
	"\\;":     1,
	"\\!":     0,
	"\\quad":  2,
	"\\qquad": 4,
}

type token struct {
	typ tokenType
	val string
}

// lexer breaks LaTeX math input into tokens. The input is treated as
// UTF-8; non-ASCII letters (Greek, etc.) are emitted as symbol tokens
// so users can paste Unicode directly instead of writing the LaTeX
// command form.
type lexer struct {
	input  string
	pos    int
	tokens []token
}

func lex(input string) []token {
	l := &lexer{input: input}
	l.run()
	return l.tokens
}

// peek returns the rune at the current position and its UTF-8 byte
// width. At EOF returns (utf8.RuneError, 0).
func (l *lexer) peek() (rune, int) {
	if l.pos >= len(l.input) {
		return utf8.RuneError, 0
	}
	return utf8.DecodeRuneInString(l.input[l.pos:])
}

// advance moves past the next rune. Returns the consumed rune.
func (l *lexer) advance() rune {
	r, size := l.peek()
	l.pos += size
	return r
}

func (l *lexer) emit(t tokenType, val string) {
	l.tokens = append(l.tokens, token{t, val})
}

func (l *lexer) run() {
	for l.pos < len(l.input) {
		r, size := l.peek()
		switch {
		case r == '$':
			// skip dollar signs (we assume we're already in math mode)
			l.pos += size
		case r == ' ', r == '\t', r == '\n', r == '\r':
			l.pos += size
		case r == '{':
			l.pos += size
			l.emit(tokLBrace, "{")
		case r == '}':
			l.pos += size
			l.emit(tokRBrace, "}")
		case r == '[':
			l.pos += size
			l.emit(tokLBracket, "[")
		case r == ']':
			l.pos += size
			l.emit(tokRBracket, "]")
		case r == '(':
			l.pos += size
			l.emit(tokLParen, "(")
		case r == ')':
			l.pos += size
			l.emit(tokRParen, ")")
		case r == '|':
			l.pos += size
			l.emit(tokPipe, "|")
		case r == '^':
			l.pos += size
			l.emit(tokCaret, "^")
		case r == '_':
			l.pos += size
			l.emit(tokUnderscore, "_")
		case r == '&':
			l.pos += size
			l.emit(tokAmpersand, "&")
		case r == '\\':
			l.lexCommand()
		case r >= '0' && r <= '9':
			l.lexNumber()
		case isOperator(r):
			l.pos += size
			l.emit(tokOperator, string(r))
		case isSymbolChar(r):
			l.pos += size
			l.emit(tokSymbol, string(r))
		default:
			l.pos += size
			l.emit(tokSymbol, string(r))
		}
	}
	l.emit(tokEOF, "")
}

func (l *lexer) lexCommand() {
	l.pos++ // consume '\' (always single byte ASCII)
	if l.pos >= len(l.input) {
		return
	}
	r, size := l.peek()
	// Handle \\ (newline in matrices)
	if r == '\\' {
		l.pos += size
		l.emit(tokNewline, "\\\\")
		return
	}
	// Handle single-char commands like \{ \} \| \,
	if !unicode.IsLetter(r) {
		l.pos += size
		switch r {
		case '{', '}':
			l.emit(tokSymbol, string(r))
		case ',':
			l.emit(tokSpace, "\\,")
		case ';':
			l.emit(tokSpace, "\\;")
		case '!':
			l.emit(tokSpace, "\\!")
		default:
			l.emit(tokSymbol, string(r))
		}
		return
	}
	// Multi-character command. LaTeX command names are ASCII letters.
	start := l.pos
	for {
		nr, nsize := l.peek()
		if nsize == 0 || !unicode.IsLetter(nr) {
			break
		}
		l.pos += nsize
	}
	cmd := l.input[start:l.pos]
	if _, ok := spaceWidths["\\"+cmd]; ok {
		l.emit(tokSpace, "\\"+cmd)
		return
	}
	if cmd == "text" || cmd == "mathrm" || cmd == "textrm" {
		l.lexTextArg(cmd)
		return
	}
	l.emit(tokCommand, cmd)
}

// lexTextArg emits a tokCommand for cmd, then — if the next non-space
// character is `{` — consumes the brace-balanced argument verbatim and
// emits its contents as a single tokText. This preserves whitespace
// inside \text{...}, which the main lexer otherwise strips.
func (l *lexer) lexTextArg(cmd string) {
	l.emit(tokCommand, cmd)
	save := l.pos
	for l.pos < len(l.input) {
		r, sz := l.peek()
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			break
		}
		l.pos += sz
	}
	if l.pos >= len(l.input) || l.input[l.pos] != '{' {
		l.pos = save
		return
	}
	l.pos++ // consume {
	start := l.pos
	depth := 1
	for l.pos < len(l.input) && depth > 0 {
		switch l.input[l.pos] {
		case '{':
			depth++
		case '}':
			depth--
		case '\\':
			// Skip an escaped char so `\}` doesn't close the group.
			if l.pos+1 < len(l.input) {
				l.pos++
			}
		}
		if depth == 0 {
			break
		}
		l.pos++
	}
	// Normalize embedded whitespace: LaTeX treats newlines and tabs
	// inside \text{...} as single spaces, and a single-cell canvas can't
	// hold a literal newline anyway — it would print as a real line break
	// mid-row and corrupt surrounding layout.
	raw := l.input[start:l.pos]
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		}
		return r
	}, raw)
	l.emit(tokText, cleaned)
	if l.pos < len(l.input) {
		l.pos++ // consume closing }
	}
}

func (l *lexer) lexNumber() {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= '0' && ch <= '9') || ch == '.' {
			l.pos++
			continue
		}
		break
	}
	l.emit(tokNumber, l.input[start:l.pos])
}

func isOperator(r rune) bool {
	if strings.ContainsRune("+-*/=<>!.,;:'\"~", r) {
		return true
	}
	return unicodeOperators[r]
}

func isSymbolChar(r rune) bool {
	return unicode.IsLetter(r)
}

// unicodeOperators recognizes math operator runes that users may paste
// directly (instead of writing the LaTeX command). Listed runes get the
// same operator-spacing treatment as their `\name` equivalents.
var unicodeOperators = map[rune]bool{
	'±': true, '∓': true, '×': true, '÷': true, '·': true,
	'≤': true, '≥': true, '≠': true, '≈': true, '≡': true,
	'∈': true, '∉': true, '⊂': true, '⊆': true, '⊃': true, '⊇': true,
	'∪': true, '∩': true,
	'→': true, '←': true, '↦': true, '⇒': true, '⇐': true, '↔': true, '⇔': true,
	'∀': true, '∃': true,
	'⟨': true, '⟩': true,
	'…': true, '⋯': true, '⋮': true, '⋱': true,
}

// parser converts tokens into an AST.
type parser struct {
	tokens []token
	pos    int
}

// parse turns a LaTeX math string into the internal AST. The shape of
// the AST is intentionally not part of the public API — callers go
// through [Render] / [Expand].
func parse(input string) (*node, error) {
	tokens := lex(input)
	p := &parser{tokens: tokens}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{tokEOF, ""}
	}
	return p.tokens[p.pos]
}

func (p *parser) next() token {
	t := p.peek()
	p.pos++
	return t
}

func (p *parser) expect(typ tokenType) (token, error) {
	t := p.next()
	if t.typ != typ {
		return t, fmt.Errorf("expected token type %d, got %d (%q)", typ, t.typ, t.val)
	}
	return t, nil
}

// parseExpr parses a sequence of atoms, collecting them into a group.
func (p *parser) parseExpr() (*node, error) {
	return p.parseExprUntil(tokEOF, tokRBrace, tokRBracket, tokAmpersand, tokNewline)
}

func (p *parser) parseExprUntil(stops ...tokenType) (*node, error) {
	var nodes []*node
	for {
		t := p.peek()
		for _, s := range stops {
			if t.typ == s {
				goto done
			}
		}
		if t.typ == tokRParen {
			goto done
		}
		// Stop on \end and \right — these terminate enclosing constructs
		if t.typ == tokCommand && (t.val == "end" || t.val == "right") {
			goto done
		}
		atom, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		if atom == nil {
			break
		}
		// Check for superscript/subscript
		atom, err = p.parsePostfix(atom)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, atom)
	}
done:
	if len(nodes) == 0 {
		return groupNode(), nil
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return groupNode(nodes...), nil
}

func (p *parser) parseAtom() (*node, error) {
	t := p.peek()
	switch t.typ {
	case tokSymbol:
		p.next()
		return symNode(t.val), nil
	case tokNumber:
		p.next()
		return numNode(t.val), nil
	case tokOperator:
		p.next()
		return opNode(t.val), nil
	case tokSpace:
		p.next()
		return spaceNode(spaceWidths[t.val]), nil
	case tokCommand:
		return p.parseCommand()
	case tokLBrace:
		return p.parseBraceGroup()
	case tokLParen:
		return p.parseDelimited("(", ")")
	case tokLBracket:
		return p.parseDelimited("[", "]")
	case tokPipe:
		if p.hasMatchingPipe() {
			p.next() // consume opening |
			node, err := p.parseExprUntil(tokEOF, tokRBrace, tokRBracket, tokAmpersand, tokNewline, tokPipe)
			if err != nil {
				return nil, err
			}
			if p.peek().typ == tokPipe {
				p.next()
			}
			return parenNode("|", "|", node), nil
		}
		p.next()
		return opNode("|"), nil
	case tokCaret:
		return nil, fmt.Errorf("missing base before %q", "^")
	case tokUnderscore:
		return nil, fmt.Errorf("missing base before %q", "_")
	default:
		return nil, nil
	}
}

func (p *parser) parseBraceGroup() (*node, error) {
	p.next() // consume {
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.expect(tokRBrace)
	if err != nil {
		return nil, fmt.Errorf("unclosed brace group: %w", err)
	}
	return node, nil
}

func (p *parser) parseDelimited(open, close string) (*node, error) {
	p.next() // consume opening delimiter
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	// consume closing delimiter
	t := p.peek()
	closeTok := delimToTok(close)
	if t.typ == closeTok {
		p.next()
	}
	return parenNode(open, close, node), nil
}

// hasMatchingPipe reports whether a closing `|` exists at the same
// brace-depth ahead of the current parser position. The current token
// is assumed to be the opening `|`. Used to disambiguate absolute-value
// delimiters (`|x|`) from the divides operator (`p | n`).
func (p *parser) hasMatchingPipe() bool {
	depth := 0
	for i := p.pos + 1; i < len(p.tokens); i++ {
		t := p.tokens[i]
		switch t.typ {
		case tokLBrace, tokLParen, tokLBracket:
			depth++
		case tokRBrace, tokRParen, tokRBracket:
			if depth == 0 {
				return false
			}
			depth--
		case tokPipe:
			if depth == 0 {
				return true
			}
		case tokAmpersand, tokNewline, tokEOF:
			if depth == 0 {
				return false
			}
		}
	}
	return false
}

func delimToTok(d string) tokenType {
	switch d {
	case ")":
		return tokRParen
	case "]":
		return tokRBracket
	case "|":
		return tokPipe
	case "}":
		return tokRBrace
	default:
		return tokEOF
	}
}

// symbolCommands maps \name → Unicode glyph for commands that produce a
// plain math symbol (Greek letters, named constants, etc).
var symbolCommands = map[string]string{
	"infty": "∞",
	// Lowercase Greek (alpha through omega)
	"alpha":   "α",
	"beta":    "β",
	"gamma":   "γ",
	"delta":   "δ",
	"epsilon": "ε",
	"zeta":    "ζ",
	"eta":     "η",
	"theta":   "θ",
	"iota":    "ι",
	"kappa":   "κ",
	"lambda":  "λ",
	"mu":      "μ",
	"nu":      "ν",
	"xi":      "ξ",
	"omicron": "ο",
	"pi":      "π",
	"rho":     "ρ",
	"sigma":   "σ",
	"tau":     "τ",
	"upsilon": "υ",
	"phi":     "φ",
	"chi":     "χ",
	"psi":     "ψ",
	"omega":   "ω",
	// Uppercase Greek (Gamma through Omega — only those distinct from
	// Latin glyphs; e.g. \Alpha is not provided since Α≡A visually)
	"Gamma":   "Γ",
	"Delta":   "Δ",
	"Theta":   "Θ",
	"Lambda":  "Λ",
	"Xi":      "Ξ",
	"Pi":      "Π",
	"Sigma":   "Σ",
	"Upsilon": "Υ",
	"Phi":     "Φ",
	"Psi":     "Ψ",
	"Omega":   "Ω",
	// Math constants & calculus
	"hbar":    "ℏ",
	"ell":     "ℓ",
	"partial": "∂",
	"nabla":   "∇",
	"prime":   "′",
}

// operatorCommands maps \name → Unicode operator glyph for commands
// that produce a binary or relational operator.
var operatorCommands = map[string]string{
	"pm":             "±",
	"mp":             "∓",
	"times":          "×",
	"div":            "÷",
	"cdot":           "·",
	"leq":            "≤",
	"le":             "≤",
	"geq":            "≥",
	"ge":             "≥",
	"neq":            "≠",
	"ne":             "≠",
	"approx":         "≈",
	"equiv":          "≡",
	"in":             "∈",
	"notin":          "∉",
	"subset":         "⊂",
	"subseteq":       "⊆",
	"cup":            "∪",
	"cap":            "∩",
	"to":             "→",
	"rightarrow":     "→",
	"leftarrow":      "←",
	"Rightarrow":     "⇒",
	"Leftarrow":      "⇐",
	"leftrightarrow": "↔",
	"Leftrightarrow": "⇔",
	"forall":         "∀",
	"exists":         "∃",
	"ast":            "∗",
	"circ":           "∘",
	"bullet":         "•",
	"star":           "⋆",
	"setminus":       "∖",
	"langle":         "⟨",
	"rangle":         "⟩",
	"mid":            "|",
	"ldots":          "…",
	"dots":           "…",
	"cdots":          "⋯",
	"vdots":          "⋮",
	"ddots":          "⋱",
}

// textCommands lists commands rendered as their plain name (sin, cos,
// log, det, …). Membership matters; the rendered text equals the key.
var textCommands = map[string]struct{}{
	"sin": {}, "cos": {}, "tan": {}, "cot": {}, "sec": {}, "csc": {},
	"arcsin": {}, "arccos": {}, "arctan": {},
	"sinh": {}, "cosh": {}, "tanh": {},
	"log": {}, "ln": {}, "exp": {},
	"det": {}, "dim": {}, "ker": {},
	"max": {}, "min": {}, "sup": {}, "inf": {},
	"mod": {}, "gcd": {}, "deg": {}, "hom": {},
}

// mathStyleTransforms maps \mathX → the text-transform function used
// to rewrite the argument's letters.
var mathStyleTransforms = map[string]func(string) string{
	"mathbf":   mathBold,
	"mathbb":   mathDoubleStruck,
	"mathcal":  mathScript,
	"mathfrak": mathFraktur,
	"mathsf":   mathSansSerif,
	"mathit":   mathItalic,
}

// accentKinds maps single-arg accent commands → the kind tag stored on
// the resulting nodeHat (used by the renderer to pick the glyph).
var accentKinds = map[string]string{
	"hat":   "hat",
	"vec":   "vec",
	"dot":   "dot",
	"ddot":  "ddot",
	"tilde": "tilde",
}

// wideAccents maps wide-accent commands → the kind tag stored on the
// resulting nodeOverline.
var wideAccents = map[string]string{
	"overline":  "",
	"bar":       "", // \bar collapses to \overline for our purposes
	"widehat":   "hat",
	"widetilde": "tilde",
}

// bigOpKinds maps big-operator commands → (nodeType, glyph).
var bigOpKinds = map[string]struct {
	typ nodeType
	sym string
}{
	"sum":  {nodeBigOp, "∑"},
	"prod": {nodeBigOp, "∏"},
	"int":  {nodeBigOp, "∫"},
	"oint": {nodeBigOp, "∮"},
}

func (p *parser) parseCommand() (*node, error) {
	t := p.next() // consume command token

	if sym, ok := symbolCommands[t.val]; ok {
		return symNode(sym), nil
	}
	if op, ok := operatorCommands[t.val]; ok {
		return opNode(op), nil
	}
	if _, ok := textCommands[t.val]; ok {
		return textNode(t.val), nil
	}
	if transform, ok := mathStyleTransforms[t.val]; ok {
		return p.parseMathStyle(transform)
	}
	if kind, ok := accentKinds[t.val]; ok {
		return p.parseAccent(kind)
	}
	if kind, ok := wideAccents[t.val]; ok {
		return p.parseOverline(kind)
	}
	if op, ok := bigOpKinds[t.val]; ok {
		return p.parseBigOp(op.typ, op.sym)
	}

	switch t.val {
	case "frac":
		return p.parseFrac()
	case "sqrt":
		return p.parseSqrt()
	case "text", "mathrm", "textrm":
		return p.parseText()
	case "left":
		return p.parseLeftRight()
	case "right":
		// should not hit this standalone; return nil
		return nil, nil
	case "lim":
		return &node{Type: nodeLim, Value: "lim"}, nil
	case "underline":
		return p.parseUnderline()
	case "overbrace":
		return p.parseBrace(nodeOverbrace, tokCaret)
	case "underbrace":
		return p.parseBrace(nodeUnderbrace, tokUnderscore)
	case "begin":
		return p.parseEnvironment()
	default:
		// Unknown command: render as text
		return textNode(t.val), nil
	}
}

func (p *parser) parseMathStyle(transform func(string) string) (*node, error) {
	arg, err := p.parseRequiredArg()
	if err != nil {
		return nil, err
	}
	return applyMathStyle(arg, transform), nil
}

func (p *parser) parseFrac() (*node, error) {
	num, err := p.parseRequiredArg()
	if err != nil {
		return nil, fmt.Errorf("frac numerator: %w", err)
	}
	den, err := p.parseRequiredArg()
	if err != nil {
		return nil, fmt.Errorf("frac denominator: %w", err)
	}
	return fracNode(num, den), nil
}

func (p *parser) parseSqrt() (*node, error) {
	// Check for optional nth root: \sqrt[n]{expr}
	if p.peek().typ == tokLBracket {
		p.next() // consume [
		n, err := p.parseExprUntil(tokRBracket)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokRBracket); err != nil {
			return nil, fmt.Errorf("sqrt index: %w", err)
		}
		expr, err := p.parseRequiredArg()
		if err != nil {
			return nil, fmt.Errorf("sqrt body: %w", err)
		}
		return nthRootNode(n, expr), nil
	}
	expr, err := p.parseRequiredArg()
	if err != nil {
		return nil, fmt.Errorf("sqrt body: %w", err)
	}
	return sqrtNode(expr), nil
}

func (p *parser) parseText() (*node, error) {
	// The lexer slurps \text{...} content verbatim into a single tokText
	// (preserving whitespace, which the main lexer otherwise strips).
	if p.peek().typ == tokText {
		t := p.next()
		return textNode(t.val), nil
	}
	// No argument: \text on its own renders as empty.
	return textNode(""), nil
}

func (p *parser) parseLeftRight() (*node, error) {
	// \left( ... \right)
	openTok := p.next()
	open := openTok.val
	if open == "." {
		open = "" // \left. means invisible delimiter
	}

	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	// expect \right
	if p.peek().typ == tokCommand && p.peek().val == "right" {
		p.next() // consume \right
	}
	closeTok := p.next()
	close := closeTok.val
	if close == "." {
		close = ""
	}

	return parenNode(open, close, node), nil
}

func (p *parser) parseBigOp(typ nodeType, sym string) (*node, error) {
	return &node{Type: typ, Value: sym}, nil
}

func (p *parser) parseOverline(kind string) (*node, error) {
	arg, err := p.parseRequiredArg()
	if err != nil {
		return nil, err
	}
	return &node{Type: nodeOverline, Value: kind, Children: []*node{arg}}, nil
}

func (p *parser) parseUnderline() (*node, error) {
	arg, err := p.parseRequiredArg()
	if err != nil {
		return nil, err
	}
	return &node{Type: nodeUnderline, Children: []*node{arg}}, nil
}

// parseBrace handles \overbrace{X}^{label} and \underbrace{X}_{label}.
// The label is consumed eagerly so it doesn't get reattached as a
// regular postfix sup/sub on the surrounding expression.
func (p *parser) parseBrace(typ nodeType, labelTok tokenType) (*node, error) {
	expr, err := p.parseRequiredArg()
	if err != nil {
		return nil, err
	}
	children := []*node{expr}
	if p.peek().typ == labelTok {
		p.next()
		label, err := p.parseRequiredArg()
		if err != nil {
			return nil, err
		}
		children = append(children, label)
	}
	return &node{Type: typ, Children: children}, nil
}

func (p *parser) parseAccent(kind string) (*node, error) {
	arg, err := p.parseRequiredArg()
	if err != nil {
		return nil, err
	}
	return &node{Type: nodeHat, Value: kind, Children: []*node{arg}}, nil
}

func (p *parser) parseEnvironment() (*node, error) {
	// parse {envname} — may be multiple tokens since lexer splits letters
	if _, err := p.expect(tokLBrace); err != nil {
		return nil, fmt.Errorf("\\begin: %w", err)
	}
	var envName string
	for p.peek().typ != tokRBrace && p.peek().typ != tokEOF {
		envName += p.next().val
	}
	if _, err := p.expect(tokRBrace); err != nil {
		return nil, fmt.Errorf("\\begin{%s}: %w", envName, err)
	}

	switch envName {
	case "pmatrix", "bmatrix", "matrix", "vmatrix", "Bmatrix", "Vmatrix", "cases":
		return p.parseMatrix(envName)
	default:
		// skip until \end{envName}
		return textNode("\\begin{" + envName + "}"), nil
	}
}

// consumeEnd parses `\end{anything}`, returning a parse error if the
// braces are malformed. The environment name itself is discarded — we
// don't verify it matches \begin, matching the lenient existing behavior.
func (p *parser) consumeEnd() error {
	p.next() // \end
	if _, err := p.expect(tokLBrace); err != nil {
		return fmt.Errorf("\\end: %w", err)
	}
	for p.peek().typ != tokRBrace && p.peek().typ != tokEOF {
		p.next()
	}
	if _, err := p.expect(tokRBrace); err != nil {
		return fmt.Errorf("\\end: %w", err)
	}
	return nil
}

func (p *parser) atEnd() bool {
	t := p.peek()
	return t.typ == tokCommand && t.val == "end"
}

func (p *parser) parseMatrix(envName string) (*node, error) {
	var rows [][]*node
	var currentRow []*node

	for {
		if p.atEnd() {
			if err := p.consumeEnd(); err != nil {
				return nil, err
			}
			break
		}
		if p.peek().typ == tokEOF {
			break
		}

		cell, err := p.parseExprUntil(tokAmpersand, tokNewline, tokEOF)
		if err != nil {
			return nil, err
		}
		currentRow = append(currentRow, cell)

		t := p.peek()
		if t.typ == tokAmpersand {
			p.next() // consume &
		} else if t.typ == tokNewline {
			p.next() // consume \\
			rows = append(rows, currentRow)
			currentRow = nil
		} else {
			if p.atEnd() {
				if err := p.consumeEnd(); err != nil {
					return nil, err
				}
			}
			break
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	open, close := matrixDelimiters(envName)
	node := &node{Type: nodeMatrix, Rows: rows, Open: open, Close: close}
	return node, nil
}

func matrixDelimiters(env string) (string, string) {
	switch env {
	case "pmatrix":
		return "(", ")"
	case "bmatrix":
		return "[", "]"
	case "vmatrix":
		return "|", "|"
	case "Bmatrix":
		return "{", "}"
	case "Vmatrix":
		return "‖", "‖"
	case "cases":
		return "{", ""
	default:
		return "", ""
	}
}

func (p *parser) parsePostfix(base *node) (*node, error) {
	// Operators can't carry scripts in LaTeX (`+^2` is a syntax error).
	// Refusing the attachment here keeps the operator visible to the
	// group-spacing rule.
	if base != nil && base.Type == nodeOperator {
		if t := p.peek(); t.typ == tokCaret || t.typ == tokUnderscore {
			return nil, fmt.Errorf("missing base before %q", t.val)
		}
		return base, nil
	}

	hasSub := false
	hasSup := false
	var sub, sup *node

	for {
		t := p.peek()
		if t.typ == tokCaret {
			if hasSup {
				return nil, fmt.Errorf("double superscript")
			}
			p.next()
			var err error
			sup, err = p.parseRequiredArg()
			if err != nil {
				return nil, err
			}
			hasSup = true
		} else if t.typ == tokUnderscore {
			if hasSub {
				return nil, fmt.Errorf("double subscript")
			}
			p.next()
			var err error
			sub, err = p.parseRequiredArg()
			if err != nil {
				return nil, err
			}
			hasSub = true
		} else {
			break
		}
	}

	if hasSub || hasSup {
		return scriptNode(base, sub, sup), nil
	}
	return base, nil
}

// parseRequiredArg parses either a brace group or a single token.
func (p *parser) parseRequiredArg() (*node, error) {
	if p.peek().typ == tokLBrace {
		return p.parseBraceGroup()
	}
	// Single token
	return p.parseAtom()
}
