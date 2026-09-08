package termtex

import (
	"strings"
	"testing"
)

// render is a test helper: Render with the given style, trailing
// whitespace trimmed, fatal on error.
func render(t *testing.T, input string, style Style) string {
	t.Helper()
	got, err := Render(input, style)
	if err != nil {
		t.Fatalf("Render(%q) error: %v", input, err)
	}
	return strings.TrimRight(got, " \n")
}

// TestAtomSpacing covers the TeX inter-atom spacing table: relations
// and binary operators get a cell on each side, punctuation a cell
// after, named operators a cell before an ordinary atom but none
// before an opening delimiter, and a leading or post-relation binary
// operator is unary.
func TestAtomSpacing(t *testing.T) {
	cases := []struct{ in, want string }{
		{`a + b = c`, "a + b = c"},
		{`x = -1`, "x = -1"},
		{`(-b)`, "(-b)"},
		{`a - b`, "a - b"},
		{`x = 1, 2, \ldots, n`, "x = 1, 2, …, n"},
		{`\{ x \in \mathbb{R} : x > 0 \}`, "{x ∈ ℝ : x > 0}"},
		{`\det(A - \lambda I) = 0`, "det(A - λI) = 0"},
		{`\sin x`, "sin x"},
		{`2\sin x \cos x`, "2 sin x cos x"},
		{`\sin(x)`, "sin(x)"},
		{`\mathrm{d}x`, "dx"},
		{`\text{if } x > 0`, "if x > 0"},
		{`n!`, "n!"},
		{`f(x)`, "f(x)"},
		{`\int x \, dF(x)`, "∫ x dF(x)"},
		{`a, b; c`, "a, b; c"},
		{`x \mid y`, "x | y"},
		{`p | n`, "p|n"},
		{`f \colon X \to Y`, "f: X → Y"},
		{`2\sqrt{2}`, "2√2"},
		{`-\sin\theta`, "-sin θ"},
		{`x = -\sin\theta`, "x = -sin θ"},
		{`\lim_{x \to c} (f)`, "lim (f)\nx→c"},
	}
	for _, c := range cases {
		if got := render(t, c.in, Style{}); got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestScriptStyleTight verifies the optional spaces vanish inside
// scripts (script style) while the mandatory ones stay.
func TestScriptStyleTight(t *testing.T) {
	got := render(t, `\sum_{i=1}^{n} a_{i,j}`, Style{})
	if strings.Contains(got, "i = 1") || strings.Contains(got, "i, j") {
		t.Errorf("script content should be tight, got:\n%s", got)
	}
}

// TestInlineStyle covers Style.Inline (TeX text style): flat
// fractions with parentheses only where needed, side limits on big
// operators, and single-row output for the common inline shapes.
func TestInlineStyle(t *testing.T) {
	inline := Style{Inline: true}
	cases := []struct{ in, want string }{
		{`\frac{a}{b}`, "a/b"},
		{`\frac{a+b}{2}`, "(a + b)/2"},
		{`\frac{1}{2a}`, "1/(2a)"},
		{`\frac{n(n+1)}{2}`, "n(n + 1)/2"},
		{`\frac{1}{1+\frac{1}{x}}`, "1/(1 + 1/x)"},
		{`\frac{\frac{a}{b}}{c}`, "(a/b)/c"},
		{`\frac{a}{b}^2`, "(a/b)²"},
		{`\sum_{i=1}^{n} x_i`, "∑ᵢ₌₁ⁿ xᵢ"},
		{`\prod_{i=1}^n i`, "∏ᵢ₌₁ⁿ i"},
		{`\int_0^1 f(x)\,dx`, "∫₀¹ f(x) dx"},
		{`x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}`, "x = (-b ± √(b² - 4ac))/(2a)"},
		{`\tfrac{a}{b}`, "a/b"},
	}
	for _, c := range cases {
		if got := render(t, c.in, inline); got != c.want {
			t.Errorf("Inline Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	// A fraction with a multi-row side still stacks inline.
	got := render(t, `\frac{\dfrac{a}{b}}{c}`, inline)
	if !strings.Contains(got, "─") {
		t.Errorf("multi-row numerator should stack, got:\n%s", got)
	}
	// \dfrac forces the stacked form even inline; \tfrac forces flat
	// even in display style.
	if got := render(t, `\dfrac{a}{b}`, inline); !strings.Contains(got, "─") {
		t.Errorf(`\dfrac should stack inline, got %q`, got)
	}
	if got := render(t, `\tfrac{a}{b}`, Style{}); got != "a/b" {
		t.Errorf(`\tfrac should be flat in display style, got %q`, got)
	}
}

// TestDisplayStyleStacks verifies the default (display) style keeps
// stacked fractions and stacked ∑ limits.
func TestDisplayStyleStacks(t *testing.T) {
	if got := render(t, `\frac{a}{b}`, Style{}); got != " a\n───\n b" {
		t.Errorf("display fraction should stack, got %q", got)
	}
	got := render(t, `\sum_{i=1}^{n} i`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 || !strings.Contains(lines[0], "n") || !strings.Contains(lines[2], "i=1") {
		t.Errorf("display sum should stack limits above/below, got:\n%s", got)
	}
}

// TestIntegralLimitsSide: integrals carry their limits as side
// scripts even in display style, as TeX does.
func TestIntegralLimitsSide(t *testing.T) {
	if got := render(t, `\int_0^1 f`, Style{}); got != "∫₀¹ f" {
		t.Errorf("integral limits should be inline side scripts, got %q", got)
	}
	// An unmappable limit stacks to the side, not above the symbol.
	got := render(t, `\int_0^\infty f`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 rows, got:\n%s", got)
	}
	if !strings.HasPrefix(lines[1], "∫") || !strings.HasPrefix(lines[0], " ∞") {
		t.Errorf("limits should sit to the right of ∫, got:\n%s", got)
	}
	// \limits forces stacking; \nolimits forces side scripts.
	got = render(t, `\int\limits_0^1 f`, Style{})
	if lines := strings.Split(got, "\n"); len(lines) != 3 || !strings.HasPrefix(lines[0], "1") {
		t.Errorf(`\int\limits should stack, got:\n%s`, got)
	}
	if got := render(t, `\sum\nolimits_{i} x`, Style{}); got != "∑ᵢ x" {
		t.Errorf(`\sum\nolimits should use side scripts, got %q`, got)
	}
	// Inline, lim writes its limit in textual form on the same row.
	if got := render(t, `\lim_{x \to 0} f`, Style{Inline: true}); got != "lim_{x→0} f" {
		t.Errorf("inline lim = %q", got)
	}
}

// TestTextualScripts: outside display style a script with no Unicode
// form is written TeX-style on one row rather than stacked.
func TestTextualScripts(t *testing.T) {
	inline := Style{Inline: true}
	cases := []struct{ in, want string }{
		{`T_c`, "T_c"},
		{`T_c + T_h`, "T_c + Tₕ"},
		{`x^{\pi}`, "x^π"},
		{`x_i^{\pi}`, "xᵢ^π"},
		{`e^{-x^2}`, "e^{-x²}"},
		{`e^{-\frac{x^2}{2}}`, "e^{-x²/2}"},
		{`\lim_{n\to\infty} a_n`, "lim_{n→∞} aₙ"},
		{`\sum_{p \text{ prime}} p`, "∑_{p prime} p"},
	}
	for _, c := range cases {
		if got := render(t, c.in, inline); got != c.want {
			t.Errorf("Inline Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// Display style keeps stacking.
	if got := render(t, `T_c`, Style{}); got != "T\n c" {
		t.Errorf("display T_c = %q", got)
	}
}

// TestScriptOrder: inline sub then sup, so a_n^2 reads as a_n squared.
func TestScriptOrder(t *testing.T) {
	if got := render(t, `a_n^2`, Style{}); got != "aₙ²" {
		t.Errorf("a_n^2 = %q, want aₙ²", got)
	}
}

// TestRadicalParens: bare symbol or number radicands take no parens;
// compound ones do; ASCII mode always uses parens.
func TestRadicalParens(t *testing.T) {
	cases := []struct {
		in    string
		style Style
		want  string
	}{
		{`\sqrt{2}`, Style{}, "√2"},
		{`\sqrt{x}`, Style{}, "√x"},
		{`\sqrt{\pi}`, Style{}, "√π"},
		{`\sqrt{x^2 + y^2}`, Style{}, "√(x² + y²)"},
		{`\sqrt{ab}`, Style{}, "√(ab)"},
		{`\sqrt{\sqrt{x}}`, Style{}, "√(√x)"},
		{`\sqrt[3]{x}`, Style{}, "³√x"},
		{`\sqrt{x}`, Style{ASCII: true}, `\(x)`},
	}
	for _, c := range cases {
		if got := render(t, c.in, c.style); got != c.want {
			t.Errorf("Render(%q, %+v) = %q, want %q", c.in, c.style, got, c.want)
		}
	}
}

// TestEquationArrays covers align/aligned/gather/array/cases and the
// bare top-level `\\` and `&` forms.
func TestEquationArrays(t *testing.T) {
	cases := []struct{ in, want string }{
		{`\begin{aligned} a &= b + c \\ &= d \end{aligned}`, "a = b + c\n  = d"},
		{`\begin{align*} x &= 1 \\ yy &= 2 \end{align*}`, " x = 1\nyy = 2"},
		{`a &= b \\ c &= d`, "a = b\nc = d"},
		{`x^2 \\ y^2`, "x²\ny²"},
		{`\begin{gather} a = b \\ cc = dd \end{gather}`, " a = b\ncc = dd"},
		{`\begin{array}{lcr} 1 & 22 & 333 \\ 4444 & 5 & 66 \end{array}`, "1     22  333\n4444  5    66"},
		{`\begin{cases} x & x \ge 0 \\ -x & x < 0 \end{cases}`, "⎧ x   x ≥ 0\n⎩ -x  x < 0"},
		{`\begin{aligned} a &= b \\ \end{aligned}`, "a = b"},
	}
	for _, c := range cases {
		if got := render(t, c.in, Style{}); got != c.want {
			t.Errorf("Render(%q) =\n%s\nwant:\n%s", c.in, got, c.want)
		}
	}
}

// TestLeftRightCommandDelimiters: \left and \right accept command
// delimiters, and \| is the double bar.
func TestLeftRightCommandDelimiters(t *testing.T) {
	cases := []struct{ in, want string }{
		{`\left\langle u, v \right\rangle`, "⟨u, v⟩"},
		{`\left\lfloor x \right\rfloor`, "⌊x⌋"},
		{`\|x\|_2`, "‖x‖₂"},
		{`\left\| x \right\|`, "‖x‖"},
		{`\lceil x \rceil`, "⌈x⌉"},
	}
	for _, c := range cases {
		if got := render(t, c.in, Style{}); got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBinom(t *testing.T) {
	got := render(t, `\binom{n}{k}`, Style{})
	if got != "⎛ n ⎞\n⎝ k ⎠" {
		t.Errorf(`\binom{n}{k} = %q`, got)
	}
}

// TestExpandInlineIsTextStyle: $...$ in markdown renders in text style
// so simple fractions and sums stay on the prose line.
func TestExpandInlineIsTextStyle(t *testing.T) {
	out := Expand(`Half is $\frac{1}{2}$ and the sum $\sum_{i=1}^n i$ here.`, Style{})
	want := "Half is 1/2 and the sum ∑ᵢ₌₁ⁿ i here."
	if out != want {
		t.Errorf("Expand inline = %q, want %q", out, want)
	}
	// Display math keeps the stacked form.
	out = Expand("$$\\frac{1}{2}$$", Style{})
	if !strings.Contains(out, "─") {
		t.Errorf("display math should stack, got %q", out)
	}
}
