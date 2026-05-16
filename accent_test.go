package termtex

import (
	"strings"
	"testing"
)

func TestOverlineUnderline(t *testing.T) {
	got, _ := Render(`\overline{x}`, Style{})
	if !strings.Contains(got, "‾") {
		t.Errorf("overline missing ‾, got: %q", got)
	}

	got, _ = Render(`\underline{x}`, Style{})
	if !strings.Contains(got, "_") {
		t.Errorf("underline missing _, got: %q", got)
	}
}

func TestAccentDot(t *testing.T) {
	got, _ := Render(`\dot{x}`, Style{})
	if !strings.ContainsRune(got, '\u0307') {
		t.Errorf(`\dot should use combining dot above, got %q`, got)
	}
}

func TestAccentDdot(t *testing.T) {
	got, _ := Render(`\ddot{x}`, Style{})
	if !strings.ContainsRune(got, '\u0308') {
		t.Errorf(`\ddot should use combining diaeresis, got %q`, got)
	}
}

func TestAccentTilde(t *testing.T) {
	got, _ := Render(`\tilde{x}`, Style{})
	if !strings.ContainsRune(got, '\u0303') {
		t.Errorf(`\tilde should use combining tilde, got %q`, got)
	}
}

func TestAccentVec(t *testing.T) {
	got, _ := Render(`\vec{v}`, Style{})
	if !strings.ContainsRune(got, '\u20D7') {
		t.Errorf(`\vec should use combining arrow, got %q`, got)
	}
}

func TestAccentHat(t *testing.T) {
	got, _ := Render(`\hat{r}`, Style{})
	if !strings.ContainsRune(got, '\u0302') {
		t.Errorf(`\hat should use combining circumflex, got %q`, got)
	}
}

func TestWideHat(t *testing.T) {
	got, _ := Render(`\widehat{abc}`, Style{})
	if !strings.Contains(got, "^^^") {
		t.Errorf(`\widehat{abc} should produce three carets, got %q`, got)
	}
}

func TestWideTilde(t *testing.T) {
	got, _ := Render(`\widetilde{xyz}`, Style{})
	if !strings.Contains(got, "~~~") {
		t.Errorf(`\widetilde{xyz} should produce three tildes, got %q`, got)
	}
}

func TestOverbrace(t *testing.T) {
	got, _ := Render(`\overbrace{a + b}^{x}`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf(`\overbrace should be 3 rows (label/brace/expr), got %d:\n%s`, len(lines), got)
	}
	if !strings.Contains(got, "╭") || !strings.Contains(got, "╮") || !strings.Contains(got, "┴") {
		t.Errorf(`\overbrace missing brace glyphs, got:\n%s`, got)
	}
}

func TestUnderbrace(t *testing.T) {
	got, _ := Render(`\underbrace{a + b}_{x}`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf(`\underbrace should be 3 rows (expr/brace/label), got %d:\n%s`, len(lines), got)
	}
	if !strings.Contains(got, "╰") || !strings.Contains(got, "╯") || !strings.Contains(got, "┬") {
		t.Errorf(`\underbrace missing brace glyphs, got:\n%s`, got)
	}
}

func TestOverbraceNoLabel(t *testing.T) {
	// A plain \overbrace{X} with no label still draws the brace.
	got, _ := Render(`\overbrace{a + b}`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf(`\overbrace without label should be 2 rows, got %d:\n%s`, len(lines), got)
	}
}
