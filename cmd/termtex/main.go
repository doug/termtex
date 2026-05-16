package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/doug/termtex"
)

// isTerminal reports whether f is connected to a terminal. Uses the
// stdlib-only ModeCharDevice check, which works on macOS, Linux, and
// Windows without pulling in x/term.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// flagSet reports whether a flag was set explicitly on the command
// line. Lets us distinguish "user passed -color=false" from "user
// didn't say anything" so the default can be auto-detected.
func flagSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

func main() {
	colorFlag := flag.Bool("color", false, "enable ANSI color output (default: auto-detect when stdout is a TTY; honors NO_COLOR)")
	italicFlag := flag.Bool("italic", false, "use Mathematical Italic Unicode (requires font support)")
	asciiFlag := flag.Bool("ascii", false, "restrict output to 7-bit ASCII")
	mdFlag := flag.Bool("md", false, "treat input as markdown; expand $...$ and $$...$$ math blocks in place")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: termtex [flags] [expression]")
		fmt.Fprintln(os.Stderr, "  Single-quote expressions so the shell doesn't strip backslashes:")
		fmt.Fprintln(os.Stderr, "       termtex '\\frac{1}{2}'")
		fmt.Fprintln(os.Stderr, "       echo '\\frac{1}{2}' | termtex")
		fmt.Fprintln(os.Stderr, "       cat doc.md | termtex -md | glow -")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	var input string
	args := flag.Args()
	switch {
	case len(args) > 0:
		input = strings.Join(args, " ")
	default:
		// No positional args and no piped stdin: print usage and exit
		// instead of blocking forever on an interactive TTY.
		if isTerminal(os.Stdin) {
			flag.Usage()
			os.Exit(1)
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
		// Preserve whitespace verbatim in markdown mode (line breaks
		// and trailing newlines matter for the surrounding markdown).
		// Trim in math mode so a trailing newline from `echo` doesn't
		// break parsing.
		if *mdFlag {
			input = string(data)
		} else {
			input = strings.TrimSpace(string(data))
		}
	}

	if input == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Color default: auto-enable when stdout is a TTY, honor NO_COLOR.
	// An explicit -color=true/false on the command line overrides.
	useColor := *colorFlag
	if !flagSet("color") {
		useColor = isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == ""
	}

	style := termtex.Style{
		Color:  useColor,
		Italic: *italicFlag,
		ASCII:  *asciiFlag,
	}

	if *mdFlag {
		fmt.Print(termtex.Expand(input, style))
		return
	}

	output, err := termtex.Render(input, style)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
}
