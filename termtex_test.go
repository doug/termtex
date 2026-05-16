package termtex

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "rewrite golden files from current output")

func TestRenderSimple(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "variable", input: "x", want: "x"},
		{name: "number", input: "42", want: "42"},
		{name: "equation", input: "a + b", want: "a + b"},
		{name: "superscript", input: "x^2", want: "x²"},
		{name: "subscript", input: "x_1", want: "x₁"},
		{name: "fraction", input: `\frac{a}{b}`, want: " a\n───\n b"},
		{name: "greek", input: `\alpha + \beta`, want: "α + β"},
		{name: "sqrt simple", input: `\sqrt{x}`, want: "√(x)"},
		{name: "sin function", input: `\sin(x)`, want: "sin (x)"},
		{name: "euler identity", input: `e^{i\pi} + 1 = 0`, want: " iπ\ne   + 1 = 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.input, Style{})
			if err != nil {
				t.Fatalf("Render(%q) error: %v", tt.input, err)
			}
			got = strings.TrimRight(got, " \n")
			want := strings.TrimRight(tt.want, " \n")
			if got != want {
				t.Errorf("Render(%q)\ngot:\n%s\nwant:\n%s", tt.input, got, want)
			}
		})
	}
}

func TestRenderComplex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "quadratic formula",
			input:    `\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
			contains: []string{"√(", "±", "─", "2a"},
		},
		{
			name:     "sum with limits",
			input:    `\sum_{i=1}^{n} i^2`,
			contains: []string{"∑", "i²", "i=1"},
		},
		{
			name:     "integral with limits",
			input:    `\int_{0}^{\infty} f(x) dx`,
			contains: []string{"∫", "∞", "0"},
		},
		{
			name:     "pythagorean",
			input:    `a^2 + b^2 = c^2`,
			contains: []string{"a²", "b²", "c²", "+", "="},
		},
		{
			name:     "fraction addition",
			input:    `\frac{1}{2} + \frac{1}{3}`,
			contains: []string{"1", "2", "3", "─", "+"},
		},
		{
			name:     "nested fraction",
			input:    `\frac{\frac{a}{b}}{c}`,
			contains: []string{"a", "b", "c", "─"},
		},
		{
			name:     "matrix",
			input:    `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`,
			contains: []string{"a", "b", "c", "d", "⎛", "⎞"},
		},
		{
			name:     "limit",
			input:    `\lim_{n \to \infty} \frac{1}{n}`,
			contains: []string{"lim", "n→∞", "─"},
		},
		{
			name:     "derivative",
			input:    `\frac{df}{dx}`,
			contains: []string{"df", "dx", "─"},
		},
		{
			name:     "trig identity",
			input:    `\sin(x)^2 + \cos(x)^2 = 1`,
			contains: []string{"sin", "cos", "²", "+", "=", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.input, Style{})
			if err != nil {
				t.Fatalf("Render(%q) error: %v", tt.input, err)
			}
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Render(%q) missing %q in:\n%s", tt.input, s, got)
				}
			}
		})
	}
}

func TestSuperscriptInline(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"x^2", "x²"},
		{"x^n", "xⁿ"},
		{"a^{10}", "a¹⁰"},
		{"x^T", "xᵀ"},
	}
	for _, tt := range tests {
		got, _ := Render(tt.input, Style{})
		got = strings.TrimRight(got, " \n")
		if got != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSuperscriptStacked(t *testing.T) {
	// Exponents with unmappable chars should stack
	got, _ := Render(`x^{\pi}`, Style{})
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Errorf("x^π should stack to multiple lines, got: %q", got)
	}
}

func TestSubscriptInline(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"x_0", "x₀"},
		{"x_i", "xᵢ"},
		{"a_{12}", "a₁₂"},
	}
	for _, tt := range tests {
		got, _ := Render(tt.input, Style{})
		got = strings.TrimRight(got, " \n")
		if got != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSqrtModes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(string) bool
		desc  string
	}{
		{
			name:  "single char parens",
			input: `\sqrt{x}`,
			check: func(s string) bool { return s == "√(x)" },
			desc:  "should be √(x)",
		},
		{
			name:  "multi char parens",
			input: `\sqrt{ab}`,
			check: func(s string) bool { return strings.Contains(s, "√(") && strings.Contains(s, ")") },
			desc:  "should use √(...)",
		},
		{
			name:  "complex parens",
			input: `\sqrt{b^2 - 4ac}`,
			check: func(s string) bool { return strings.Contains(s, "√(") },
			desc:  "should use √(...) for single-line complex",
		},
		{
			name:  "multi-line tall parens",
			input: `\sqrt{\frac{a}{b}}`,
			check: func(s string) bool { return strings.Contains(s, "⎛") && strings.Contains(s, "⎝") },
			desc:  "should use tall parens for multi-line content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.input, Style{})
			if err != nil {
				t.Fatal(err)
			}
			got = strings.TrimRight(got, " \n")
			if !tt.check(got) {
				t.Errorf("Render(%q) %s, got:\n%s", tt.input, tt.desc, got)
			}
		})
	}
}

func TestRenderItalic(t *testing.T) {
	style := Style{Italic: true}
	got, _ := Render("x", style)
	got = strings.TrimRight(got, " \n")
	if got != "𝑥" {
		t.Errorf("expected 𝑥, got %q", got)
	}
}

func TestRenderColor(t *testing.T) {
	style := Style{Color: true}
	got, _ := Render("x + 1", style)
	if !strings.Contains(got, "\033[") {
		t.Error("expected ANSI escape codes in colored output")
	}
	if !strings.Contains(got, "\033[0m") {
		t.Error("expected ANSI reset in colored output")
	}
}

func TestMeasureBox(t *testing.T) {
	tests := []struct {
		input    string
		height   int
		baseline int
	}{
		{`\frac{a}{b}`, 3, 1},
		{`x^2`, 1, 0},                   // inline superscript
		{`x`, 1, 0},                     // single symbol
		{`\frac{\frac{a}{b}}{c}`, 5, 3}, // nested fraction
	}
	for _, tt := range tests {
		node, _ := parse(tt.input)
		box := measure(node, newRenderCtx(Style{}))
		if box.Height != tt.height {
			t.Errorf("measure(%q) height = %d, want %d", tt.input, box.Height, tt.height)
		}
		if box.Baseline != tt.baseline {
			t.Errorf("measure(%q) baseline = %d, want %d", tt.input, box.Baseline, tt.baseline)
		}
	}
}

func TestBigOpLimits(t *testing.T) {
	// Sum with limits should stack vertically
	got, _ := Render(`\sum_{i=1}^{n}`, Style{})
	lines := strings.Split(strings.TrimRight(got, " \n"), "\n")
	if len(lines) != 3 {
		t.Errorf("sum with limits should be 3 lines, got %d:\n%s", len(lines), got)
	}
}

func TestMatrixTypes(t *testing.T) {
	tests := []struct {
		env  string
		open string
	}{
		{"pmatrix", "⎛"},
		{"bmatrix", "⎡"},
	}
	for _, tt := range tests {
		input := `\begin{` + tt.env + `} a & b \\ c & d \end{` + tt.env + `}`
		got, err := Render(input, Style{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, tt.open) {
			t.Errorf("%s should contain %q, got:\n%s", tt.env, tt.open, got)
		}
	}
}

func TestSpacingCompactLimits(t *testing.T) {
	// i=1 in limits should NOT have spaces around =
	got, _ := Render(`\sum_{i=1}^{n}`, Style{})
	if strings.Contains(got, "i = 1") {
		t.Errorf("limit subscript should be compact, got:\n%s", got)
	}
}

func TestSpacingNormalEquation(t *testing.T) {
	// Top-level = should have spaces
	got, _ := Render(`a + b = c`, Style{})
	if !strings.Contains(got, " = ") {
		t.Errorf("equation should space =, got: %q", got)
	}
	if !strings.Contains(got, " + ") {
		t.Errorf("equation should space +, got: %q", got)
	}
}

func TestTextOperatorSpacing(t *testing.T) {
	got, _ := Render(`\sin(x)`, Style{})
	if !strings.Contains(got, "sin ") {
		t.Errorf("expected space after sin, got: %q", got)
	}
}

// TestRenderGolden compares the rendered output of a curated set of
// expressions against committed golden files in testdata/. Run with
// `go test -update` to regenerate the goldens after intentional
// rendering changes. Tests catch full-layout regressions that
// strings.Contains can't (column alignment, baseline shifts, etc).
func TestRenderGolden(t *testing.T) {
	cases := map[string]string{
		"euler-identity":  `e^{i\pi} + 1 = 0`,
		"quadratic":       `\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
		"sum-of-squares":  `\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}`,
		"gaussian":        `\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`,
		"binomial":        `(x + y)^n = \sum_{k=0}^{n} \frac{n!}{k!(n-k)!} x^k y^{n-k}`,
		"taylor":          `f(x) = \sum_{n=0}^{\infty} \frac{f^{(n)}(a)}{n!} (x - a)^n`,
		"matrix-2x2":      `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`,
		"matrix-inverse":  `A^{-1} = \frac{1}{ad - bc} \begin{bmatrix} d & -b \\ -c & a \end{bmatrix}`,
		"length-contract": `L = L_0 \sqrt{1 - \frac{v^2}{c^2}}`,
		"hat":             `\hat{r}`,
		"vec":             `\vec{v}`,
		"carnot":          `\eta = 1 - \frac{T_c}{T_h}`,
		"unicode-input":   `α + β = γ`,
		"abs-value":       `|x + y|`,
		"divides":         `\prod_{p | n} f(p)`,
		"cube-root":       `\sqrt[3]{x + y}`,
	}
	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Render(input, Style{})
			if err != nil {
				t.Fatalf("Render(%q) error: %v", input, err)
			}
			path := filepath.Join(dir, name+".golden")
			if *updateGolden {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v (run with -update to create)", path, err)
			}
			if got != string(want) {
				t.Errorf("Render(%q) golden mismatch (-want +got):\n--- want\n%s\n--- got\n%s", input, want, got)
			}
		})
	}
}
