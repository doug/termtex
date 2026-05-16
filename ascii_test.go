package termtex

import (
	"strings"
	"testing"
)

func TestASCIIPureBytes(t *testing.T) {
	style := Style{}
	style.ASCII = true
	inputs := []string{
		`x`,
		`a + b = c`,
		`x^2 + y^2 = z^2`,
		`\frac{a}{b}`,
		`\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
		`\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}`,
		`\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`,
		`e^{i\pi} + 1 = 0`,
		`\alpha + \beta + \gamma = \pi`,
		`\nabla \cdot E = \frac{\rho}{\epsilon_0}`,
		`\begin{pmatrix} a & b \\ c & d \end{pmatrix}`,
		`\begin{bmatrix} 1 & 0 \\ 0 & 1 \end{bmatrix}`,
		`x \in S \subseteq \mathbb{R}`,
		`A \to B \Rightarrow C`,
		`\forall x \exists y`,
		`\hat{x} + \bar{y} + \vec{z}`,
		`\overline{a + b}`,
		`\lim_{x \to 0} \frac{\sin x}{x}`,
		`\sqrt[3]{x + y}`,
		`\mathbb{R}^n`,
		`\mathcal{L}(\mathbf{x})`,
		`\mathfrak{g}_{\mathsf{abc}}`,
		`\dot{x} + \ddot{y} + \tilde{z}`,
		`\vec{v} = \widehat{abc}`,
		`\widetilde{X}`,
		`\overbrace{a + b}^{label}`,
		`\underbrace{1 + 2}_{sum}`,
	}
	for _, in := range inputs {
		got, err := Render(in, style)
		if err != nil {
			t.Errorf("Render(%q) error: %v", in, err)
			continue
		}
		for i, r := range got {
			if r > 0x7F {
				t.Errorf("ASCII render of %q produced non-ASCII rune %U at byte offset %d:\n%s", in, r, i, got)
				break
			}
		}
	}
}

func TestASCIIDisablesItalic(t *testing.T) {
	// Italic uses Mathematical Italic Unicode; ASCII mode must skip it.
	style := Style{}
	style.ASCII = true
	style.Italic = true
	got, err := Render(`x + y`, style)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	for _, r := range got {
		if r > 0x7F {
			t.Errorf("ASCII+Italic produced non-ASCII rune %U: %q", r, got)
			break
		}
	}
}

func TestASCIIBasicShape(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, err := Render(`a + b = c`, style)
	if err != nil {
		t.Fatal(err)
	}
	got = strings.TrimRight(got, " \n")
	if got != "a + b = c" {
		t.Errorf("ASCII a+b=c got %q", got)
	}
}

func TestASCIIFractionBar(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\frac{a}{b}`, style)
	if !strings.Contains(got, "---") {
		t.Errorf("ASCII fraction should use --- bar, got:\n%s", got)
	}
}

func TestASCIIGreekMapping(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\alpha + \beta`, style)
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") {
		t.Errorf("ASCII Greek should spell out names, got: %q", got)
	}
}

func TestAccentASCIIDegrades(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\dot{x} + \ddot{y} + \vec{v}`, style)
	for _, r := range got {
		if r > 0x7F {
			t.Errorf("ASCII accents produced non-ASCII rune %U: %q", r, got)
			break
		}
	}
}

func TestOverbraceASCIIDegrades(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\overbrace{a + b}^{x}`, style)
	for _, r := range got {
		if r > 0x7F {
			t.Errorf("ASCII overbrace produced non-ASCII rune %U: %q", r, got)
			break
		}
	}
	if !strings.Contains(got, "^") || !strings.Contains(got, "+") {
		t.Errorf("ASCII overbrace missing expected glyphs, got:\n%s", got)
	}
}

func TestMathStyleASCIIDegrades(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\mathbb{R}`, style)
	got = strings.TrimRight(got, " \n")
	if got != "R" {
		t.Errorf(`ASCII \mathbb{R} = %q, want "R"`, got)
	}
	got, _ = Render(`\mathcal{L}`, style)
	got = strings.TrimRight(got, " \n")
	if got != "L" {
		t.Errorf(`ASCII \mathcal{L} = %q, want "L"`, got)
	}
}

func TestASCIIBigOpMapping(t *testing.T) {
	style := Style{}
	style.ASCII = true
	got, _ := Render(`\sum_{i=1}^{n}`, style)
	if !strings.Contains(got, "Sum") {
		t.Errorf("ASCII sum should be 'Sum', got:\n%s", got)
	}
}
