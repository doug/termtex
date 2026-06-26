package termtex

import (
	"strings"
	"testing"
)

func TestExpandDisplayMath(t *testing.T) {
	md := "Before\n\n$$\\frac{1}{2}$$\n\nAfter"
	out := Expand(md, Style{})

	if !strings.Contains(out, "─") {
		t.Errorf("display math not rendered:\n%s", out)
	}
	if !strings.Contains(out, "Before") || !strings.Contains(out, "After") {
		t.Errorf("surrounding text lost:\n%s", out)
	}
	// Should be wrapped in fenced code block
	if !strings.Contains(out, "```") {
		t.Errorf("display math should be in fenced code block:\n%s", out)
	}
}

func TestExpandInlineMath(t *testing.T) {
	md := "The formula $a^2 + b^2 = c^2$ is famous."
	out := Expand(md, Style{})

	// Simple inline should render on one line with unicode superscripts
	if !strings.Contains(out, "a²") {
		t.Errorf("inline math not rendered:\n%s", out)
	}
	if !strings.Contains(out, "is famous") {
		t.Errorf("surrounding text lost:\n%s", out)
	}
}

func TestExpandPreservesNonMath(t *testing.T) {
	md := "# Hello World\n\nThis has no math."
	out := Expand(md, Style{})
	if out != md {
		t.Errorf("non-math markdown should be unchanged:\ngot:  %q\nwant: %q", out, md)
	}
}

func TestExpandEscapedDollar(t *testing.T) {
	md := `The price is \$5 and \$10.`
	out := Expand(md, Style{})
	// Should not try to parse \$ as math
	if strings.Contains(out, "─") {
		t.Errorf("escaped dollars should not be parsed as math:\n%s", out)
	}
}

// TestExpandProseDollarsPassThrough covers the three Pandoc
// rules in the regex-replacement path. Currency-style prose must not
// be mistaken for inline math.
func TestExpandProseDollarsPassThrough(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"I'll give you $3 dollars if you give me $5.", "$3 dollars if you give me $5"},
		{"It costs $20,000 and $30,000.", "$20,000 and $30,000"},
		{"That book is $5 now.", "$5 now"},
		{"With $ x $ spacing.", "$ x $"},
	}
	for _, c := range cases {
		out := Expand(c.in, Style{})
		if !strings.Contains(out, c.want) {
			t.Errorf("%q: expected %q in output, got:\n%s", c.in, c.want, out)
		}
		if strings.Contains(out, "─") {
			t.Errorf("%q: prose was typeset as math:\n%s", c.in, out)
		}
	}
}

func TestExpandMathInsideCodeBlocksUnchanged(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "inline code",
			in:   "This is `code with $a^2 + b^2 = c^2$` and math $a^2 + b^2 = c^2$.",
			// The second one should be typeset (we can check for unicode/fraction, or just check that they are different).
			// Let's check that the code span remains exactly as is.
			want: "This is `code with $a^2 + b^2 = c^2$` and math ",
		},
		{
			name: "fenced code block",
			in:   "Before\n```go\nx := $a^2$\n```\nAfter $a^2$.",
			want: "Before\n```go\nx := $a^2$\n```\nAfter ",
		},
		{
			name: "multiple backticks inline code",
			in:   "This is ``code with $a^2$`` and math $a^2$.",
			want: "This is ``code with $a^2$`` and math ",
		},
		{
			name: "backslash inside code span",
			in:   "This is inline code: `a \\` b $a^2$`",
			want: "This is inline code: `a \\` b a²",
		},
		{
			name: "HTML code tag",
			in:   "Before <code>x := $a^2$</code> After $a^2$.",
			want: "Before <code>x := $a^2$</code> After ",
		},
		{
			name: "HTML pre tag with attributes",
			in:   "Before <pre class=\"go\">x := $a^2$</pre> After $a^2$.",
			want: "Before <pre class=\"go\">x := $a^2$</pre> After ",
		},
		{
			name: "HTML tag case insensitivity",
			in:   "Before <CODE>x := $a^2$</CODE> After $a^2$.",
			want: "Before <CODE>x := $a^2$</CODE> After ",
		},
		{
			name: "HTML tag not code",
			in:   "Before <div class=\"code\">x := $a^2$</div> After $a^2$.",
			want: "Before <div class=\"code\">x := a²</div> After a²",
		},
		{
			name: "tilde fenced code block",
			in:   "Before\n~~~go\nx := $a^2$\n~~~\nAfter $a^2$.",
			want: "Before\n~~~go\nx := $a^2$\n~~~\nAfter ",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Expand(c.in, Style{})
			if !strings.HasPrefix(out, c.want) {
				t.Errorf("Expected output to start with %q, got:\n%s", c.want, out)
			}
			// In each case, the second math expression should be typeset (which contains superscripts like a²).
			if !strings.Contains(out, "a²") {
				t.Errorf("Expected outside math to be expanded, got:\n%s", out)
			}
			// In each case, the inside math expression should NOT be typeset (meaning the literal string "$a^2$" or "$a^2$'" remains).
			// Wait, the first one is `code with $a^2 ...$` so checking for `$a^2$` is correct.
			if c.name == "inline code" && !strings.Contains(out, "$a^2 + b^2 = c^2$") {
				t.Errorf("Expected inline code math to remain unchanged, got:\n%s", out)
			}
			if c.name == "fenced code block" && !strings.Contains(out, "$a^2$") {
				t.Errorf("Expected fenced code block math to remain unchanged, got:\n%s", out)
			}
		})
	}
}

