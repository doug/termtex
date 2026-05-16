package termtex

import (
	"strings"
	"testing"
)

// TestRegression covers bugs surfaced by the May 2026 layout/parser
// audit. Each case asserts a single observable property, kept loose
// enough to survive minor cosmetic changes in the renderer.

func TestRegressionNthRootTallIndex(t *testing.T) {
	// `\sqrt[\frac{1}{2}]{x}` previously rendered the fraction index
	// into a 1-row canvas, clipping the bar and denominator.
	got, err := Render(`\sqrt[\frac{1}{2}]{x}`, Style{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "─") {
		t.Errorf("fraction bar of nth-root index missing:\n%s", got)
	}
	if !strings.Contains(got, "2") {
		t.Errorf("denominator of nth-root index missing:\n%s", got)
	}
}

func TestRegressionInlineScriptOnFrac(t *testing.T) {
	// `\frac{a}{b}_x` used to put ₓ next to the numerator. The
	// subscript should appear on the denominator row.
	got, _ := Render(`\frac{a}{b}_x`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected 3-row layout, got:\n%s", got)
	}
	if !strings.Contains(lines[len(lines)-1], "x") {
		t.Errorf("subscript should be on the denominator row, got:\n%s", got)
	}
}

func TestRegressionTextWhitespace(t *testing.T) {
	// Embedded newlines and tabs inside \text{...} used to be stamped
	// as literal cells in the canvas, corrupting downstream layout.
	got, _ := Render("\\text{foo\nbar}", Style{})
	if strings.Contains(got, "\n") {
		t.Errorf("text should normalize newline to space, got: %q", got)
	}
	got, _ = Render("\\text{foo\tbar}", Style{})
	if strings.Contains(got, "\t") {
		t.Errorf("text should normalize tab to space, got: %q", got)
	}
}

func TestRegressionEmptyMatrix(t *testing.T) {
	// measureMatrix reserves 2 cells for an empty pmatrix; the renderer
	// must paint the delimiters to fill them.
	got, _ := Render(`\begin{pmatrix}\end{pmatrix}`, Style{})
	if !strings.Contains(got, "(") || !strings.Contains(got, ")") {
		t.Errorf("empty pmatrix should still show delimiters, got: %q", got)
	}
}

func TestRegressionDoubleScriptError(t *testing.T) {
	// `x_{a}^{b}_{c}` previously dropped `_{c}` silently. Surface
	// the malformed input as a parse error.
	if _, err := Render(`x_{a}^{b}_{c}`, Style{}); err == nil {
		t.Errorf("double subscript should be a parse error")
	}
	if _, err := Render(`x^{a}^{b}`, Style{}); err == nil {
		t.Errorf("double superscript should be a parse error")
	}
}

func TestRegressionOrphanScriptError(t *testing.T) {
	// A leading `^x` / `_x` used to abandon the rest of the input.
	for _, in := range []string{`^x`, `_x`, `^x y`, `_x y`} {
		if _, err := Render(in, Style{}); err == nil {
			t.Errorf("orphan script %q should be a parse error", in)
		}
	}
}

func TestRegressionEscapedDisplayMath(t *testing.T) {
	// `\$$x$$` is a backslash-escaped dollar in markdown; display
	// math should NOT trigger.
	out := Expand(`\$$x$$`, Style{})
	if strings.Contains(out, "```") {
		t.Errorf("escaped \\$$ should not trigger display math, got:\n%s", out)
	}
}

func TestRegressionMatrixBaselineAlign(t *testing.T) {
	// In a row containing both a fraction and a plain symbol, the
	// plain symbol should align with the fraction's bar (its baseline),
	// not with the numerator.
	got, _ := Render(`\begin{pmatrix} \frac{a}{b} & x \\ y & z \end{pmatrix}`, Style{})
	lines := strings.Split(got, "\n")
	// Find the line that contains the fraction bar and assert `x` is on it.
	for _, line := range lines {
		if strings.Contains(line, "─") {
			if !strings.Contains(line, "x") {
				t.Errorf("plain symbol should baseline-align with frac bar; got line %q in:\n%s", line, got)
			}
			return
		}
	}
	t.Errorf("expected a fraction bar in matrix output:\n%s", got)
}

func TestRegressionEmptyArgNoPhantomRow(t *testing.T) {
	// `\overbrace{x}^{}` used to produce a leading blank row from
	// the empty group's height. Likewise `\hat{}` and `\sum_{}^{}`.
	for _, in := range []string{`\overbrace{x}^{}`, `\sum_{}^{}`, `\hat{}`} {
		got, _ := Render(in, Style{})
		if strings.HasPrefix(got, "\n") {
			t.Errorf("%q produced leading blank row:\n%s", in, got)
		}
	}
}

func TestRegressionOperatorPostfixError(t *testing.T) {
	// `y + ^2 + z` used to silently produce `y+² + z` (operator
	// spacing broken). It's actually malformed; surface the error.
	if _, err := Render(`y + ^2 + z`, Style{}); err == nil {
		t.Errorf("operator with postfix script should be a parse error")
	}
}
