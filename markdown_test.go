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
