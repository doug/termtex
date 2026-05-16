package termtex

import (
	"strings"
	"testing"
)

// TestGreekLetters verifies the full alpha-through-omega and
// Gamma-through-Omega ranges claimed in the README. Catches the gap
// where individual letters were missing from the symbolCommands table.
func TestGreekLetters(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Lowercase
		{`\alpha`, "α"}, {`\beta`, "β"}, {`\gamma`, "γ"}, {`\delta`, "δ"},
		{`\epsilon`, "ε"}, {`\zeta`, "ζ"}, {`\eta`, "η"}, {`\theta`, "θ"},
		{`\iota`, "ι"}, {`\kappa`, "κ"}, {`\lambda`, "λ"}, {`\mu`, "μ"},
		{`\nu`, "ν"}, {`\xi`, "ξ"}, {`\omicron`, "ο"}, {`\pi`, "π"},
		{`\rho`, "ρ"}, {`\sigma`, "σ"}, {`\tau`, "τ"}, {`\upsilon`, "υ"},
		{`\phi`, "φ"}, {`\chi`, "χ"}, {`\psi`, "ψ"}, {`\omega`, "ω"},
		// Uppercase (only those distinct from Latin glyphs)
		{`\Gamma`, "Γ"}, {`\Delta`, "Δ"}, {`\Theta`, "Θ"}, {`\Lambda`, "Λ"},
		{`\Xi`, "Ξ"}, {`\Pi`, "Π"}, {`\Sigma`, "Σ"}, {`\Upsilon`, "Υ"},
		{`\Phi`, "Φ"}, {`\Psi`, "Ψ"}, {`\Omega`, "Ω"},
		// Math constants
		{`\infty`, "∞"}, {`\hbar`, "ℏ"}, {`\partial`, "∂"}, {`\nabla`, "∇"},
	}
	for _, tt := range tests {
		got, _ := Render(tt.input, Style{})
		got = strings.TrimRight(got, " \n")
		if got != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`\pm`, "±"},
		{`\times`, "×"},
		{`\cdot`, "·"},
		{`\leq`, "≤"},
		{`\geq`, "≥"},
		{`\neq`, "≠"},
		{`\approx`, "≈"},
		{`\in`, "∈"},
		{`\rightarrow`, "→"},
		{`\Rightarrow`, "⇒"},
		{`\forall`, "∀"},
		{`\exists`, "∃"},
		{`\partial`, "∂"},
		{`\nabla`, "∇"},
	}
	for _, tt := range tests {
		got, _ := Render(tt.input, Style{})
		got = strings.TrimRight(got, " \n")
		if got != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMathFunctions(t *testing.T) {
	fns := []string{"sin", "cos", "tan", "log", "ln", "exp", "det", "min", "max"}
	for _, fn := range fns {
		got, _ := Render(`\`+fn, Style{})
		got = strings.TrimRight(got, " \n")
		if got != fn {
			t.Errorf(`Render(\%s) = %q, want %q`, fn, got, fn)
		}
	}
}

func TestParseErrors(t *testing.T) {
	// These should not panic
	inputs := []string{
		"",
		"{}",
		"{",
		`\frac{a}`,
		`\sqrt`,
		`^2`,
		`_i`,
		`\unknown`,
	}
	for _, input := range inputs {
		Render(input, Style{}) // should not panic
	}
}

// TestUnicodeInput verifies the rune-based lexer accepts pasted
// Unicode (Greek letters, math symbols) in addition to LaTeX commands.
func TestUnicodeInput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`α + β = γ`, "α + β = γ"},
		{`σ²`, "σ²"},
		{`π × ε`, "π × ε"},
	}
	for _, tt := range tests {
		got, err := Render(tt.input, Style{})
		if err != nil {
			t.Errorf("Render(%q) error: %v", tt.input, err)
			continue
		}
		got = strings.TrimRight(got, " \n")
		if got != tt.want {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestPipeAsInfixOperator verifies a bare `|` between operands is
// parsed as the divides operator, not as an opening abs-value
// delimiter that would synthesize a closing pipe.
func TestPipeAsInfixOperator(t *testing.T) {
	got, _ := Render(`p | n`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "p|n" {
		t.Errorf("p|n should render as p|n, got %q", got)
	}
}

// TestPipeAsAbsValue verifies matched `|...|` is parsed as an
// absolute-value delimiter pair.
func TestPipeAsAbsValue(t *testing.T) {
	got, _ := Render(`|x|`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "|x|" {
		t.Errorf("|x| should render as |x|, got %q", got)
	}
}
