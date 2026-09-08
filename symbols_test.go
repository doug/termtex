package termtex

import (
	"strings"
	"testing"
)

// TestSymbolCoverage: every command in the tables renders as a single
// glyph, and a representative sample maps to the expected character.
func TestSymbolCoverage(t *testing.T) {
	sample := map[string]string{
		`\varepsilon`: "ε", `\vartheta`: "ϑ", `\varphi`: "φ", `\emptyset`: "∅",
		`\aleph`: "ℵ", `\Re`: "ℜ", `\Im`: "ℑ", `\neg`: "¬", `\top`: "⊤",
		`\angle`: "∠", `\oplus`: "⊕", `\otimes`: "⊗", `\land`: "∧", `\lor`: "∨",
		`\iff`: "⟺", `\implies`: "⟹", `\perp`: "⊥", `\parallel`: "∥",
		`\propto`: "∝", `\sim`: "∼", `\simeq`: "≃", `\cong`: "≅", `\ll`: "≪",
		`\models`: "⊨", `\vdash`: "⊢", `\mapsto`: "↦", `\hookrightarrow`: "↪",
		`\uparrow`: "↑", `\therefore`: "∴", `\lt`: "<", `\gt`: ">",
		`\dagger`: "†", `\star`: "⋆", `\nmid`: "∤", `\ni`: "∋",
		`\lbrace`: "{", `\rbrace`: "}", `\degree`: "°", `\Alpha`: "A",
	}
	for in, want := range sample {
		if got := render(t, in, Style{}); got != want {
			t.Errorf("Render(%q) = %q, want %q", in, got, want)
		}
	}
	// Nothing in the tables falls through to the unknown-command path.
	for name := range symbolCommands {
		if got := render(t, `\`+name, Style{}); got == name {
			t.Errorf(`\%s rendered as its own name`, name)
		}
	}
	for name := range operatorCommands {
		if got := render(t, `\`+name, Style{}); got == name {
			t.Errorf(`\%s rendered as its own name`, name)
		}
	}
}

// TestSymbolSpacing: the new glyphs pick up the right atom class.
func TestSymbolSpacing(t *testing.T) {
	cases := []struct{ in, want string }{
		{`a \oplus b`, "a ⊕ b"},
		{`A \land B \lor \neg C`, "A ∧ B ∨ ¬C"},
		{`a \iff b \implies c`, "a ⟺ b ⟹ c"},
		{`a \perp b \parallel c`, "a ⊥ b ∥ c"},
		{`x \sim y`, "x ∼ y"},
		{`\therefore x`, "∴ x"},
		{`\lfloor x \rfloor`, "⌊x⌋"},
		{`\mathbb{R} \ni x`, "ℝ ∋ x"},
		{`a \dagger b`, "a † b"},
	}
	for _, c := range cases {
		if got := render(t, c.in, Style{}); got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNot(t *testing.T) {
	cases := []struct{ in, want string }{
		{`a \not= b`, "a ≠ b"},
		{`x \not\in S`, "x ∉ S"},
		{`p \not\mid q`, "p ∤ q"},
		{`a \not\subset b`, "a ⊄ b"},
		{`a \not\equiv b`, "a ≢ b"},
		{`a \not\perp b`, "a ⊥̸ b"}, // no precomposed form: combining overlay
	}
	for _, c := range cases {
		if got := render(t, c.in, Style{}); got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestModOperators(t *testing.T) {
	if got := render(t, `a \equiv b \pmod{n}`, Style{}); got != "a ≡ b (mod n)" {
		t.Errorf(`\pmod = %q`, got)
	}
	if got := render(t, `a \bmod n`, Style{}); got != "a mod n" {
		t.Errorf(`\bmod = %q`, got)
	}
}

func TestOperatorName(t *testing.T) {
	if got := render(t, `\operatorname{argmax}_x f(x)`, Style{}); got != "argmaxₓ f(x)" {
		t.Errorf(`\operatorname = %q`, got)
	}
	got := render(t, `\operatorname*{argmax}_x f(x)`, Style{})
	if got != "argmax f(x)\n  x" {
		t.Errorf(`\operatorname* should stack its limit, got %q`, got)
	}
	// Named limit operators stack in display style, side-script inline.
	got = render(t, `\max_{x} f(x)`, Style{})
	if !strings.HasPrefix(got, "max f(x)\n") {
		t.Errorf(`\max_{x} should stack in display style, got %q`, got)
	}
	if got := render(t, `\max_{x} f(x)`, Style{Inline: true}); got != "maxₓ f(x)" {
		t.Errorf(`\max_{x} inline = %q`, got)
	}
	if got := render(t, `\sgn(x)`, Style{}); got != "sgn(x)" {
		t.Errorf(`\sgn = %q`, got)
	}
}

func TestSubstack(t *testing.T) {
	got := render(t, `\sum_{\substack{i<j \\ i+j=n}} a_i`, Style{})
	want := "  ∑   aᵢ\n i<j\ni+j=n"
	if got != want {
		t.Errorf("substack =\n%s\nwant:\n%s", got, want)
	}
}

func TestOverUnderSet(t *testing.T) {
	if got := render(t, `a \overset{?}{=} b`, Style{}); got != "  ?\na = b" {
		t.Errorf(`\overset = %q`, got)
	}
	if got := render(t, `\underset{x}{\max}`, Style{}); got != "max\n x" {
		t.Errorf(`\underset = %q`, got)
	}
}

func TestXArrow(t *testing.T) {
	got := render(t, `A \xrightarrow{f} B`, Style{})
	if got != "   f\nA ──→ B" {
		t.Errorf(`\xrightarrow = %q`, got)
	}
	got = render(t, `A \xleftarrow[g]{h} B`, Style{})
	if got != "   h\nA ←── B\n   g" {
		t.Errorf(`\xleftarrow[g]{h} = %q`, got)
	}
	got = render(t, `A \xrightarrow{long} B`, Style{})
	if got != "   long\nA ─────→ B" {
		t.Errorf(`\xrightarrow{long} = %q`, got)
	}
	if got := render(t, `\overrightarrow{AB}`, Style{}); got != "─→\nAB" {
		t.Errorf(`\overrightarrow = %q`, got)
	}
}

func TestMoreAccents(t *testing.T) {
	cases := []struct {
		in   string
		mark rune
	}{
		{`\bar{x}`, '̄'},
		{`\breve{a}`, '̆'},
		{`\check{c}`, '̌'},
		{`\acute{e}`, '́'},
		{`\grave{e}`, '̀'},
		{`\mathring{A}`, '̊'},
	}
	for _, c := range cases {
		got := render(t, c.in, Style{})
		if !strings.ContainsRune(got, c.mark) || strings.Contains(got, "\n") {
			t.Errorf("%s should be a single row with combining %U, got %q", c.in, c.mark, got)
		}
	}
	// Multi-char base falls back to a mark row above.
	if got := render(t, `\bar{xy}`, Style{}); got != "‾‾\nxy" {
		t.Errorf(`\bar{xy} = %q`, got)
	}
}

func TestFontsAndDecorations(t *testing.T) {
	if got := render(t, `\boldsymbol{x}`, Style{}); got != "𝐱" {
		t.Errorf(`\boldsymbol = %q`, got)
	}
	if got := render(t, `\mathtt{a}`, Style{}); got != "𝚊" {
		t.Errorf(`\mathtt = %q`, got)
	}
	if got := render(t, `\cancel{x}`, Style{}); got != "x̸" {
		t.Errorf(`\cancel = %q`, got)
	}
	// The overlay is zero-width: it must not eat the spacing cell.
	if got := render(t, `\cancel{x} + y`, Style{}); got != "x̸ + y" {
		t.Errorf(`\cancel{x} + y = %q`, got)
	}
	if got := render(t, `\cancel{x}`, Style{ASCII: true}); got != "x" {
		t.Errorf(`ASCII \cancel = %q`, got)
	}
	// Bold and italic cover Greek too.
	if got := render(t, `\boldsymbol{\alpha}`, Style{}); got != "𝛂" {
		t.Errorf(`\boldsymbol{\alpha} = %q`, got)
	}
	if got := render(t, `\boldsymbol{\Omega}`, Style{}); got != "𝛀" {
		t.Errorf(`\boldsymbol{\Omega} = %q`, got)
	}
	if got := render(t, `\mathit{\pi}`, Style{}); got != "𝜋" {
		t.Errorf(`\mathit{\pi} = %q`, got)
	}
	if got := render(t, `\boldsymbol{\alpha}`, Style{ASCII: true}); got != "alpha" {
		t.Errorf(`ASCII \boldsymbol{\alpha} = %q`, got)
	}
}

func TestSizesSpacesPhantom(t *testing.T) {
	if got := render(t, `\bigl( x \bigr)`, Style{}); got != "(x)" {
		t.Errorf(`\bigl( = %q`, got)
	}
	if got := render(t, `\phantom{xx}y`, Style{}); got != "  y" {
		t.Errorf(`\phantom = %q`, got)
	}
	if got := render(t, `a~b`, Style{}); got != "a b" {
		t.Errorf("tie = %q", got)
	}
	if got := render(t, `a\ b`, Style{}); got != "a b" {
		t.Errorf(`\  = %q`, got)
	}
	if got := render(t, `\textbf{bold text}`, Style{}); got != "bold text" {
		t.Errorf(`\textbf = %q`, got)
	}
}

// TestASCIIExtendedSymbols: the new glyphs all have ASCII fallbacks
// rather than degrading to '?'.
func TestASCIIExtendedSymbols(t *testing.T) {
	for name, glyph := range extraOperators {
		got := render(t, `\`+name, Style{ASCII: true})
		if strings.Contains(got, "?") {
			t.Errorf(`ASCII \%s (%s) = %q`, name, glyph, got)
		}
	}
	for name, glyph := range extraSymbols {
		got := render(t, `\`+name, Style{ASCII: true})
		if strings.Contains(got, "?") {
			t.Errorf(`ASCII \%s (%s) = %q`, name, glyph, got)
		}
	}
}
