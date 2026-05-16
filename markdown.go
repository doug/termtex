package termtex

import "strings"

func isMathSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// isEscaped reports whether md[i] is preceded by an odd number of
// backslashes (i.e. `\$` is escaped, but `\\$` is not).
func isEscaped(md string, i int) bool {
	n := 0
	for k := i - 1; k >= 0 && md[k] == '\\'; k-- {
		n++
	}
	return n%2 == 1
}

// Expand scans markdown text for math delimiters and replaces
// them with termtex-rendered output. Display math ($$...$$) becomes
// fenced code blocks that glamour preserves verbatim. Inline math ($...$)
// is rendered inline. Pass [Style]{} for the package default.
//
// The result can be passed to glamour or any other terminal markdown
// renderer.
func Expand(md string, style Style) string {
	// Process display math first ($$...$$) to avoid matching inside them
	md = processDisplayMath(md, style)
	// Then inline math ($...$)
	md = processInlineMath(md, style)
	return md
}

// processDisplayMath replaces `$$...$$` blocks (possibly spanning
// multiple lines) with fenced code blocks containing the rendered
// output. A backslash-escaped `\$$...$$` is preserved as literal, to
// match the inline-math path.
func processDisplayMath(md string, style Style) string {
	var sb strings.Builder
	sb.Grow(len(md))
	i := 0
	for i < len(md) {
		if i+1 >= len(md) || md[i] != '$' || md[i+1] != '$' {
			sb.WriteByte(md[i])
			i++
			continue
		}
		if isEscaped(md, i) {
			sb.WriteByte(md[i])
			i++
			continue
		}
		end := strings.Index(md[i+2:], "$$")
		if end < 0 {
			sb.WriteByte(md[i])
			i++
			continue
		}
		end += i + 2
		expr := strings.TrimSpace(md[i+2 : end])
		if expr == "" {
			sb.WriteString(md[i : end+2])
			i = end + 2
			continue
		}
		rendered, err := Render(expr, style)
		if err != nil {
			sb.WriteString(md[i : end+2])
			i = end + 2
			continue
		}
		sb.WriteString("\n```\n")
		sb.WriteString(rendered)
		sb.WriteString("\n```\n")
		i = end + 2
	}
	return sb.String()
}

// processInlineMath replaces `$...$` math runs with their rendered
// output. The three Pandoc rules apply: the opener can't be followed
// by whitespace, the closer can't be preceded by whitespace, and the
// closer can't be followed by an ASCII digit. Together, those make
// "I'll give you $3 dollars if you give me $5" pass through as prose.
// A backslash-escaped `$` is treated as a literal.
func processInlineMath(md string, style Style) string {
	var sb strings.Builder
	sb.Grow(len(md))
	i := 0
	for i < len(md) {
		if md[i] != '$' {
			sb.WriteByte(md[i])
			i++
			continue
		}
		// Backslash-escaped `$` is a literal.
		if isEscaped(md, i) {
			sb.WriteByte(md[i])
			i++
			continue
		}
		// `$$` (display math) was handled earlier; leave any leftovers alone.
		if i+1 < len(md) && md[i+1] == '$' {
			sb.WriteByte(md[i])
			sb.WriteByte(md[i+1])
			i += 2
			continue
		}
		// Rule 1: opener must be followed by a non-whitespace character.
		if i+1 >= len(md) || isMathSpace(md[i+1]) {
			sb.WriteByte(md[i])
			i++
			continue
		}
		// Scan for the closer on the same line. Skip escaped `\$`
		// (LaTeX literal dollar) inside the expression.
		end := -1
		for j := i + 2; j < len(md); j++ {
			if md[j] == '\n' {
				break
			}
			if md[j] == '\\' && j+1 < len(md) {
				j++ // skip the escaped char
				continue
			}
			if md[j] != '$' {
				continue
			}
			// Rule 2: closer can't be preceded by whitespace.
			if isMathSpace(md[j-1]) {
				continue
			}
			// Rule 3: closer can't be followed by an ASCII digit.
			if j+1 < len(md) && md[j+1] >= '0' && md[j+1] <= '9' {
				continue
			}
			end = j
			break
		}
		if end < 0 {
			sb.WriteByte(md[i])
			i++
			continue
		}
		expr := strings.TrimSpace(md[i+1 : end])
		if expr == "" {
			sb.WriteByte(md[i])
			i++
			continue
		}
		rendered, err := Render(expr, style)
		if err != nil {
			sb.WriteByte(md[i])
			i++
			continue
		}
		if strings.Contains(rendered, "\n") {
			sb.WriteString("\n\n```\n")
			sb.WriteString(rendered)
			sb.WriteString("\n```\n\n")
		} else {
			sb.WriteString(rendered)
		}
		i = end + 1
	}
	return sb.String()
}
