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
//
// Within a single Expand call, the same math expression is only
// rendered once — repeated occurrences hit a per-call cache. This
// matters in practice: a typical AI-conversation transcript or paper
// uses the same handful of expressions (\alpha, x_i, f(x), ...) many
// times.
func Expand(md string, style Style) string {
	// Sentinel: empty value cached on parse error so a malformed
	// repeat-expression still costs only one lookup, not a re-parse.
	cache := make(map[string]string)
	tryRender := func(expr string) (string, bool) {
		if r, ok := cache[expr]; ok {
			return r, r != ""
		}
		r, err := Render(expr, style)
		if err != nil {
			cache[expr] = ""
			return "", false
		}
		cache[expr] = r
		return r, true
	}

	var sb strings.Builder
	sb.Grow(len(md))
	i := 0
	n := len(md)
	inCode := false
	codeOpenerLength := 0
	codeOpenerChar := byte(0)
	htmlCloser := ""

	for i < n {
		if inCode {
			j := strings.IndexByte(md[i:], codeOpenerChar)
			if j < 0 {
				sb.WriteString(md[i:])
				break
			}
			j += i
			if j > i {
				sb.WriteString(md[i:j])
			}
			runStart := j
			runLen := 0
			for j < n && md[j] == codeOpenerChar {
				runLen++
				j++
			}
			sb.WriteString(md[runStart:j])
			if runLen == codeOpenerLength {
				inCode = false
			}
			i = j
			continue
		}

		if htmlCloser != "" {
			closerLen := len(htmlCloser)
			j := -1
			searchStart := i
			for {
				pos := strings.Index(md[searchStart:], "</")
				if pos < 0 {
					break
				}
				matchIdx := searchStart + pos
				if matchIdx+closerLen <= n && strings.EqualFold(md[matchIdx:matchIdx+closerLen], htmlCloser) {
					j = matchIdx
					break
				}
				searchStart = matchIdx + 2
			}
			if j < 0 {
				sb.WriteString(md[i:])
				break
			}
			sb.WriteString(md[i : j+closerLen])
			i = j + closerLen
			htmlCloser = ""
			continue
		}

		idx, char := findNextOpener(md, i)
		if idx < 0 {
			sb.WriteString(md[i:])
			break
		}
		j := i + idx
		if j > i {
			sb.WriteString(md[i:j])
		}
		i = j

		if char == '`' || char == '~' {
			runStart := i
			runLen := 0
			for i < n && md[i] == char {
				runLen++
				i++
			}
			if isEscaped(md, runStart) {
				sb.WriteString(md[runStart:i])
				continue
			}
			inCode = true
			codeOpenerLength = runLen
			codeOpenerChar = char
			sb.WriteString(md[runStart:i])
			continue
		}

		if char == '<' {
			lower := strings.ToLower(md[i:])
			var closer string
			if strings.HasPrefix(lower, "<code") && (len(lower) == 5 || lower[5] == '>' || lower[5] == ' ' || lower[5] == '\t' || lower[5] == '\n' || lower[5] == '\r') {
				closer = "</code>"
			} else if strings.HasPrefix(lower, "<pre") && (len(lower) == 4 || lower[4] == '>' || lower[4] == ' ' || lower[4] == '\t' || lower[4] == '\n' || lower[4] == '\r') {
				closer = "</pre>"
			}

			if closer != "" {
				tagEnd := strings.IndexByte(md[i:], '>')
				if tagEnd >= 0 {
					sb.WriteString(md[i : i+tagEnd+1])
					i = i + tagEnd + 1
					htmlCloser = closer
					continue
				}
			}

			sb.WriteByte('<')
			i++
			continue
		}

		// Escaped `\$` → literal dollar; emit and advance.
		if isEscaped(md, i) {
			sb.WriteByte('$')
			i++
			continue
		}

		// Display math: $$...$$
		if i+1 < n && md[i+1] == '$' {
			end := strings.Index(md[i+2:], "$$")
			if end < 0 {
				// No closer: emit the `$$` literally and continue.
				sb.WriteString(md[i : i+2])
				i += 2
				continue
			}
			end += i + 2
			expr := strings.TrimSpace(md[i+2 : end])
			if expr == "" {
				sb.WriteString(md[i : end+2])
				i = end + 2
				continue
			}
			rendered, ok := tryRender(expr)
			if !ok {
				sb.WriteString(md[i : end+2])
				i = end + 2
				continue
			}
			sb.WriteString("\n```\n")
			sb.WriteString(rendered)
			sb.WriteString("\n```\n")
			i = end + 2
			continue
		}

		// Inline math: $...$ on a single line, with the Pandoc rules.
		// Rule 1: opener must be followed by non-whitespace.
		if i+1 >= n || isMathSpace(md[i+1]) {
			sb.WriteByte('$')
			i++
			continue
		}
		end := findInlineClose(md, i)
		if end < 0 {
			sb.WriteByte('$')
			i++
			continue
		}
		expr := strings.TrimSpace(md[i+1 : end])
		if expr == "" {
			sb.WriteByte('$')
			i++
			continue
		}
		rendered, ok := tryRender(expr)
		if !ok {
			sb.WriteByte('$')
			i++
			continue
		}
		if strings.IndexByte(rendered, '\n') >= 0 {
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

// findInlineClose scans for the closing `$` of an inline math run that
// starts at md[i]. Returns the index of the closer, or -1 if none on
// the same line. Mirrors the Pandoc rules: skip escaped `\$`, closer
// can't be preceded by whitespace, closer can't be followed by an
// ASCII digit.
func findInlineClose(md string, i int) int {
	for j := i + 2; j < len(md); j++ {
		c := md[j]
		if c == '\n' {
			return -1
		}
		if c == '\\' && j+1 < len(md) {
			j++ // skip the escaped char
			continue
		}
		if c != '$' {
			continue
		}
		if isMathSpace(md[j-1]) {
			continue
		}
		if j+1 < len(md) && md[j+1] >= '0' && md[j+1] <= '9' {
			continue
		}
		return j
	}
	return -1
}

// findNextOpener finds the index of the first occurrence of '$', '`', '<', or '~' in md[i:].
// Returns the index (relative to i) and the character found. If none are found, returns -1, 0.
func findNextOpener(md string, i int) (int, byte) {
	n := len(md) - i
	minPos := n
	var foundChar byte

	if nextDollar := strings.IndexByte(md[i:i+minPos], '$'); nextDollar >= 0 {
		minPos = nextDollar
		foundChar = '$'
	}
	if nextBacktick := strings.IndexByte(md[i:i+minPos], '`'); nextBacktick >= 0 {
		minPos = nextBacktick
		foundChar = '`'
	}
	if nextHTML := strings.IndexByte(md[i:i+minPos], '<'); nextHTML >= 0 {
		minPos = nextHTML
		foundChar = '<'
	}
	if nextTilde := strings.IndexByte(md[i:i+minPos], '~'); nextTilde >= 0 {
		minPos = nextTilde
		foundChar = '~'
	}

	if minPos == n {
		return -1, 0
	}
	return minPos, foundChar
}
