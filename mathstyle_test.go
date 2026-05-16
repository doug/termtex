package termtex

import (
	"strings"
	"testing"
)

func TestMathStyleDoubleStruck(t *testing.T) {
	got, _ := Render(`\mathbb{R}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "ℝ" {
		t.Errorf(`\mathbb{R} = %q, want "ℝ"`, got)
	}
	got, _ = Render(`\mathbb{N}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "ℕ" {
		t.Errorf(`\mathbb{N} = %q, want "ℕ"`, got)
	}
	// Letter without a BMP override → fall back to U+1D538 + offset
	got, _ = Render(`\mathbb{A}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "𝔸" {
		t.Errorf(`\mathbb{A} = %q, want "𝔸"`, got)
	}
}

func TestMathStyleScript(t *testing.T) {
	got, _ := Render(`\mathcal{L}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "ℒ" {
		t.Errorf(`\mathcal{L} = %q, want "ℒ"`, got)
	}
	got, _ = Render(`\mathcal{A}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "𝒜" {
		t.Errorf(`\mathcal{A} = %q, want "𝒜"`, got)
	}
}

func TestMathStyleBold(t *testing.T) {
	got, _ := Render(`\mathbf{x}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "𝐱" {
		t.Errorf(`\mathbf{x} = %q, want "𝐱"`, got)
	}
}

func TestMathStyleFraktur(t *testing.T) {
	got, _ := Render(`\mathfrak{g}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "𝔤" {
		t.Errorf(`\mathfrak{g} = %q, want "𝔤"`, got)
	}
	got, _ = Render(`\mathfrak{R}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "ℜ" {
		t.Errorf(`\mathfrak{R} = %q, want "ℜ"`, got)
	}
}

func TestMathStyleSansSerif(t *testing.T) {
	got, _ := Render(`\mathsf{X}`, Style{})
	got = strings.TrimRight(got, " \n")
	if got != "𝖷" {
		t.Errorf(`\mathsf{X} = %q, want "𝖷"`, got)
	}
}

func TestMathStyleNested(t *testing.T) {
	// Style applies to letters inside a structured arg.
	got, _ := Render(`\mathbb{R}^n`, Style{})
	got = strings.TrimRight(got, " \n")
	if !strings.Contains(got, "ℝ") {
		t.Errorf(`expected styled R in %q`, got)
	}
}
