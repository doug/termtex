package termtex

import "testing"

// FuzzParse verifies the parser doesn't panic on arbitrary input.
// Seeds cover common LaTeX shapes plus a handful of edge cases the
// parser has historically mishandled (unmatched delimiters, nested
// commands, Unicode, deep nesting).
func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"x",
		`\frac{a}{b}`,
		`\sqrt{x}`,
		`\sqrt[3]{x}`,
		`x^2 + y^2 = z^2`,
		`\sum_{i=1}^{n} i^2`,
		`\int_{0}^{\infty} e^{-x^2} dx`,
		`\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
		`e^{i\pi} + 1 = 0`,
		`|x|`,
		`p | n`,
		`\prod_{p | n} f(p)`,
		`\hat{r}`,
		`\widehat{abc}`,
		`\begin{bmatrix} a & b \\ c & d \end{bmatrix}`,
		`\left( \frac{a}{b} \right)`,
		`α + β = γ`, // Unicode input
		`{`,
		`{{`,
		`}{`,
		`\frac{`,
		`\sqrt`,
		`\unknowncommand{x}`,
		`x_{}^{}`,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Cap input length so the fuzzer doesn't waste time on
		// pathologically large strings; the interesting bugs are at
		// short inputs anyway.
		if len(input) > 1024 {
			return
		}
		// Render with each style mode so style-specific render paths
		// (ASCII fallback, italic codepoint substitution, color escapes)
		// get exercised too. Parse errors are fine; panics are not.
		for _, style := range []Style{
			{},
			{ASCII: true},
			{Italic: true},
			{Color: true},
		} {
			_, _ = Render(input, style)
		}
	})
}
