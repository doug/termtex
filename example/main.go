// Example shows the practical termtex + glamour pipeline: pre-process
// markdown to expand $...$ and $$...$$ math into rendered Unicode
// blocks, then pass the result to glamour for terminal styling.
//
// The goldmark subpackage (a goldmark Extender that adds MathInline /
// MathBlock nodes) isn't shown here. Glamour owns its goldmark
// instance and doesn't expose an extension hook, so for glamour users
// the preprocess path is the only option. The goldmark extension is
// for downstream projects building their own terminal renderer on
// top of goldmark.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/doug/termtex"
)

const sampleMarkdown = `# The Quadratic Formula

Given a quadratic equation $ax^2 + bx + c = 0$, the solutions are:

$$\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$$

## Euler's Identity

The most beautiful equation in mathematics: $e^{i\pi} + 1 = 0$

## Sum Formula

$$\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}$$

## Pythagorean Theorem

For a right triangle: $a^2 + b^2 = c^2$
`

func main() {
	// Step 1: expand math delimiters into rendered termtex output.
	processed := termtex.Expand(sampleMarkdown, termtex.Style{})

	// Step 2: pass the rewritten markdown to glamour as usual.
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "glamour: %v\n", err)
		os.Exit(1)
	}

	out, err := r.Render(processed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
