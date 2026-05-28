package termtex

import (
	"fmt"
	"strings"
	"testing"
)

// largeMarkdown builds a representative markdown document with prose,
// inline math, and display math. n controls the number of "sections" —
// each section ~1 KB of content.
func largeMarkdown(n int) string {
	section := `# Section heading

The quadratic equation $ax^2 + bx + c = 0$ has solutions given by:

$$x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$$

Some prose with prices like $5 and $10,000 that should not be math.
A second paragraph talks about Euler's identity $e^{i\pi} + 1 = 0$ and
the sum $\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}$ as classics.

Inline references: $\alpha$, $\beta$, $\gamma$, $\delta$, $\theta$,
$\phi$, $\psi$, $\omega$, $\nabla \cdot \vec{E} = \rho/\epsilon_0$.

$$\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}$$

Some more prose. The Pythagorean theorem $a^2 + b^2 = c^2$ holds for
right triangles. A matrix example:

$$\begin{pmatrix} 1 & 2 \\ 3 & 4 \end{pmatrix}$$

End of section.

`
	var b strings.Builder
	b.Grow(len(section) * n)
	for i := 0; i < n; i++ {
		b.WriteString(section)
	}
	return b.String()
}

func BenchmarkExpandSmall(b *testing.B) {
	md := largeMarkdown(1)
	b.SetBytes(int64(len(md)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Expand(md, Style{})
	}
}

func BenchmarkExpandLarge(b *testing.B) {
	md := largeMarkdown(100)
	b.SetBytes(int64(len(md)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Expand(md, Style{})
	}
}

func BenchmarkExpandHuge(b *testing.B) {
	md := largeMarkdown(1000)
	b.SetBytes(int64(len(md)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Expand(md, Style{})
	}
}

// BenchmarkExpandProse measures the scan-only path: a long markdown
// document with no math content at all. Should be ~memcpy-fast.
func BenchmarkExpandProse(b *testing.B) {
	const para = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.

`
	md := strings.Repeat(para, 500)
	b.SetBytes(int64(len(md)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Expand(md, Style{})
	}
}

func BenchmarkRenderQuadratic(b *testing.B) {
	const expr = `\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Render(expr, Style{})
	}
}

// largeMarkdownVaried builds a markdown document whose math expressions
// vary across sections — defeating any per-expression cache so we
// measure the worst case: every expression is unique.
func largeMarkdownVaried(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, `# Section %d

Body text with a per-section formula $f_{%d}(x) = x^{%d} + %d$ inline.

$$\sum_{i=1}^{%d} i^{%d} = \frac{n(n+%d)}{%d}$$

More prose. Inline: $\alpha_{%d}$ and $\beta_{%d}$ and a fraction
$\frac{%d}{%d}$ closing the paragraph.

`, i, i, i%9+1, i, i+1, i%5+1, i, i+1, i, i, i%7, i%11+1)
	}
	return b.String()
}

func BenchmarkExpandLargeVaried(b *testing.B) {
	md := largeMarkdownVaried(100)
	b.SetBytes(int64(len(md)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Expand(md, Style{})
	}
}

func BenchmarkRenderInline(b *testing.B) {
	const expr = `a^2 + b^2 = c^2`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Render(expr, Style{})
	}
}
