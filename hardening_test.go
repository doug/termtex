package termtex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The whole suite runs with strict canvas bounds: any write outside
// the measured box panics instead of being clipped, so a measure pass
// and a render pass that disagree fail loudly.
func init() {
	strictCanvas = true
}

// TestGalleryGolden renders testdata/gallery.md through Expand — every
// demo equation as display math plus inline uses — and compares the
// whole document against testdata/gallery.golden. Regenerate with
// `go test -update` after intentional rendering changes.
func TestGalleryGolden(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "gallery.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := Expand(string(src), Style{})
	path := filepath.Join("testdata", "gallery.golden")
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
		// Report the first differing line to keep the failure readable.
		gl := strings.Split(got, "\n")
		wl := strings.Split(string(want), "\n")
		for i := range gl {
			if i >= len(wl) || gl[i] != wl[i] {
				t.Fatalf("gallery golden mismatch at line %d:\n--- want\n%s\n--- got\n%s", i+1, strings.Join(wl[max(0, i-3):min(len(wl), i+4)], "\n"), strings.Join(gl[max(0, i-3):min(len(gl), i+4)], "\n"))
			}
		}
		t.Fatalf("gallery golden mismatch (got is shorter)")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestPerScriptStacking: a script that cannot be inlined stacks on its
// own; siblings keep their inline form.
func TestPerScriptStacking(t *testing.T) {
	got := render(t, `T_c + T_h`, Style{})
	if got != "T  + Tₕ\n c" {
		t.Errorf("T_c + T_h = %q", got)
	}
	// Nested compound scripts stack rather than flattening ambiguously.
	got = render(t, `e^{x^2}`, Style{})
	if got != " x²\ne" {
		t.Errorf("e^{x^2} = %q", got)
	}
	if got := render(t, `x^{n+1}`, Style{}); got != "xⁿ⁺¹" {
		t.Errorf("flat compound script should still inline, got %q", got)
	}
}

// TestStyleSwitch: \displaystyle inside inline math restores stacked
// layout for the rest of the list; \textstyle does the reverse.
func TestStyleSwitch(t *testing.T) {
	inline := Style{Inline: true}
	got := render(t, `\displaystyle\sum_{i=1}^{n} i`, inline)
	if lines := strings.Split(got, "\n"); len(lines) != 3 {
		t.Errorf(`\displaystyle should stack limits inline, got %q`, got)
	}
	if got := render(t, `a + \displaystyle\frac{1}{2}`, inline); !strings.Contains(got, "─") {
		t.Errorf(`\displaystyle\frac should stack inline, got %q`, got)
	}
	if got := render(t, `\textstyle\frac{1}{2}`, Style{}); got != "1/2" {
		t.Errorf(`\textstyle\frac in display style = %q`, got)
	}
	// The switch applies only to the rest of its own group.
	if got := render(t, `{\textstyle\frac{1}{2}} + \frac{1}{2}`, Style{}); !strings.Contains(got, "1/2 +") || !strings.Contains(got, "─") {
		t.Errorf("switch should not leak out of its group, got:\n%s", got)
	}
}

// TestLineBreaking: expressions wider than Style.Width break before
// relations, then binary operators, with indented continuation rows.
func TestLineBreaking(t *testing.T) {
	in := `a + b + c = d + e + f = g + h + i`
	if got := render(t, in, Style{Width: 80}); strings.Contains(got, "\n") {
		t.Errorf("fits, should not break: %q", got)
	}
	got := render(t, in, Style{Width: 20})
	want := "a + b + c\n    = d + e + f\n    = g + h + i"
	if got != want {
		t.Errorf("break at relations:\n%s\nwant:\n%s", got, want)
	}
	// No relation narrow enough: fall back to binary operators.
	got = render(t, `a + b + c + d + e + f + g + h`, Style{Width: 14})
	want = "a + b + c + d\n    + e + f\n    + g + h"
	if got != want {
		t.Errorf("break at binary operators:\n%s\nwant:\n%s", got, want)
	}
	// Multi-row pieces keep their shape.
	got = render(t, `x = \frac{a}{b} + \frac{c}{d} = y`, Style{Width: 10})
	if !strings.Contains(got, "─") || !strings.Contains(got, "\n    = y") {
		t.Errorf("multi-row break:\n%s", got)
	}
	// Inline math in markdown never breaks.
	out := Expand("$a + b = c + d = e$", Style{Width: 5})
	if strings.Contains(out, "```") {
		t.Errorf("inline math should ignore Width, got %q", out)
	}
}
