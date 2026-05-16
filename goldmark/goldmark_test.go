package goldmark

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestMathBlockRendering(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))

	source := []byte("# Title\n\n$$\\frac{1}{2}$$\n\nText after.\n")
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// Should contain rendered fraction (─ bar)
	if !strings.Contains(out, "─") {
		t.Errorf("math block not rendered:\n%s", out)
	}
	// Should contain surrounding markup
	if !strings.Contains(out, "Title") {
		t.Errorf("title lost:\n%s", out)
	}
}

func TestMathInlineRendering(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))

	source := []byte("The value $x^2$ is positive.\n")
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// Should contain unicode superscript
	if !strings.Contains(out, "x²") {
		t.Errorf("inline math not rendered, got:\n%s", out)
	}
}

func TestNoMathPassthrough(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))

	source := []byte("Just a paragraph.\n")
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Just a paragraph.") {
		t.Errorf("plain text lost:\n%s", out)
	}
}

// TestProseDollarSignsPassThrough covers the three Pandoc rules:
// opener-not-followed-by-space, closer-not-preceded-by-space, and
// closer-not-followed-by-digit. The canonical currency examples must
// render verbatim rather than being typeset as inline math.
func TestProseDollarSignsPassThrough(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	cases := []struct {
		in   string
		want string // substring that must appear verbatim in the output
	}{
		{"I'll give you $3 dollars if you give me $5.", "$3 dollars if you give me $5"},
		{"It costs $20,000 and $30,000.", "$20,000 and $30,000"},
		{"That book is $5 now.", "$5 now"},
		{"With $ x $ spacing.", "$ x $"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := md.Convert([]byte(c.in+"\n"), &buf); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, c.want) {
			t.Errorf("%q: expected %q in output, got:\n%s", c.in, c.want, got)
		}
	}
}

// TestBackslashDollarEscape verifies Pandoc's documented escape: an
// authored `\$` is a literal dollar sign and never starts math.
func TestBackslashDollarEscape(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	source := []byte(`The price is \$5 and \$10, not math.` + "\n")
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Literal dollar signs should survive; rendered math would have
	// emitted Unicode super/subscripts or box-drawing.
	if !strings.Contains(out, "$5") || !strings.Contains(out, "$10") {
		t.Errorf("escaped dollars lost:\n%s", out)
	}
	if strings.Contains(out, "─") {
		t.Errorf("escaped dollars triggered math rendering:\n%s", out)
	}
}

// TestTexDelimiters covers KaTeX/Pandoc's `\(...\)` (inline) and
// `\[...\]` (display) — the delimiters that exist precisely to avoid
// the $-ambiguity.
func TestTexDelimiters(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(NewMathExtension()))
	cases := []struct {
		in   string
		want string
	}{
		{`The value \(x^2\) is positive.`, "x²"},
		{`See \[a^2 + b^2 = c^2\] above.`, "a²"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := md.Convert([]byte(c.in+"\n"), &buf); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, c.want) {
			t.Errorf("%q: expected %q in output, got:\n%s", c.in, c.want, got)
		}
	}
}
