package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ErrInterrupt is returned by ReadLine when the user presses Ctrl-C.
var ErrInterrupt = errors.New("interrupt")

type LineEditor struct {
	fd     int
	in     *bufio.Reader
	out    io.Writer
	prompt string
}

func NewLineEditor(prompt string) *LineEditor {
	return &LineEditor{
		fd:     int(os.Stdin.Fd()),
		in:     bufio.NewReader(os.Stdin),
		out:    os.Stdout,
		prompt: prompt,
	}
}

// ReadLine reads one line. It puts the terminal in raw mode for the duration of
// the line and restores it before returning
func (e *LineEditor) ReadLine() (string, error) {
	state, err := term.MakeRaw(e.fd)
	if err != nil {
		return e.readLineCooked() // not a terminal (e.g. piped input)
	}
	defer term.Restore(e.fd, state)

	e.write(e.prompt)

	var line []rune
	tabCount := 0 // consecutive tabs with no completion progress
	for {
		r, _, err := e.in.ReadRune()
		if err != nil {
			if err == io.EOF {
				return "", io.EOF
			}
			return "", err
		}

		switch {
		case r == '\r' || r == '\n': // Enter
			e.write("\r\n")
			return string(line), nil

		case r == 3: // Ctrl-C
			e.write("\r\n")
			return "", ErrInterrupt

		case r == 4: // Ctrl-D
			if len(line) == 0 {
				e.write("\r\n")
				return "", io.EOF
			}
			// non-empty line: ignored for now

		case r == 127 || r == 8: // Backspace / DEL
			if len(line) > 0 {
				line = line[:len(line)-1]
				e.write("\b \b") // erase the last rune
			} else {
				e.bell() // nothing to delete
			}
			tabCount = 0

		case r == '\t': // tab autocomplete
			if !strings.Contains(string(line), " ") {
				// Autocomplete for commands (the first word).
				line, tabCount = e.autocompleteCommand(line, tabCount)
			} else {
				// Autocomplete for arguments: complete the last field as a path.
				line, tabCount = e.autocompleteArgument(line, tabCount)
			}

		case r == 27: // ESC: swallow arrow keys etc. so they don't corrupt the line
			e.consumeEscape()

		case r >= 32: // printable
			line = append(line, r)
			e.write(string(r))
			tabCount = 0
		}
	}
}

// autocompleteCommand handles a Tab press. It mirrors bash: extend to the longest common
// prefix when that makes progress, ring the bell when it can't, and list all
// candidates on the second consecutive bell. Returns the (possibly extended)
// line and the new consecutive-tab count.
func (e *LineEditor) autocompleteCommand(line []rune, tabCount int) ([]rune, int) {
	prefix := string(line)

	// We only complete the command (the first word). Once there's a space,
	// there's nothing here to complete.
	if prefix == "" || strings.Contains(prefix, " ") {
		e.bell()
		return line, 0
	}

	matches := cmdTrie.getMatches(prefix)

	switch len(matches) {
	case 0:
		// No matches: ring the bell, reprint the prompt, and clear the line.
		e.bell()
		e.write("\r\n")
		e.write(e.prompt)
		return nil, 0

	case 1:
		completed := matches[0] + " " // unique match → trailing space for args
		e.write(completed[len(prefix):])
		return []rune(completed), 0

	default:
		// All matches share at least `prefix`; extend to their common prefix.
		if lcp := longestCommonPrefix(matches); len(lcp) > len(prefix) {
			e.write(lcp[len(prefix):])
			return []rune(lcp), 0
		}
		// No further common prefix: bell once, list on the next tab.
		if tabCount == 0 {
			e.bell()
			return line, 1
		}
		e.write("\r\n" + strings.Join(matches, "  ") + "\r\n")
		e.write(e.prompt + prefix)
		return line, 0
	}
}

// Same as above but for directories
func (e *LineEditor) autocompleteArgument(line []rune, tabCount int) ([]rune, int) {
	parts := strings.Split(string(line), " ")
	last := len(parts) - 1
	field := parts[last]

	var dirDisplay, statDir, prefix string
	if i := strings.LastIndex(field, "/"); i >= 0 { // check if the user is trying to complete a path within a directory
		dirDisplay = field[:i+1] // what's shown on the line
		prefix = field[i+1:]     // the partial name to complete within that directory
		statDir = field[:i]      // the directory to stat for completion
		if statDir == "" {
			statDir = "/" // leading slash: read the filesystem root
		}
	} else {
		prefix = field
		statDir = "."
	}

	matches := getPathMatches(statDir, prefix, false)

	switch len(matches) {
	case 0:
		// No matches: bell, but leave the line intact (unlike command
		// completion, clearing a half-typed argument would be hostile).
		e.bell()
		return line, 0

	case 1:
		// A directory gets a trailing "/" so you can tab into it; anything else
		// gets a space to end the field. isDir follows symlinks, matching bash.
		suffix := " "
		if isDir(dirDisplay + matches[0]) {
			suffix = "/"
		}
		autocompleted := matches[0] + suffix

		e.write(autocompleted[len(prefix):])
		parts[last] = dirDisplay + autocompleted
		return []rune(strings.Join(parts, " ")), 0

	default:
		// All matches share at least `prefix`; extend to their common prefix.
		if lcp := longestCommonPrefix(matches); len(lcp) > len(prefix) {
			e.write(lcp[len(prefix):])
			parts[last] = dirDisplay + lcp
			return []rune(strings.Join(parts, " ")), 0
		}
		// No further common prefix: bell once, list on the next tab.
		if tabCount == 0 {
			e.bell()
			return line, 1
		}
		// List the candidates, marking directories with a trailing "/".
		labels := make([]string, len(matches))
		for i, name := range matches {
			labels[i] = name
			if isDir(dirDisplay + name) {
				labels[i] += "/"
			}
		}
		e.write("\r\n" + strings.Join(labels, "  ") + "\r\n")
		e.write(e.prompt + string(line))
		return line, 0
	}
}

// consumeEscape discards an escape sequence (e.g. an arrow key) using only the
// bytes already buffered, so a lone ESC keypress doesn't block waiting for more.
func (e *LineEditor) consumeEscape() {
	for e.in.Buffered() > 0 {
		b, err := e.in.ReadByte()
		if err != nil {
			return
		}
		// A CSI/SS3 sequence ends on a byte in 0x40-0x7e (other than the
		// '['/'O' introducer).
		if b >= 0x40 && b <= 0x7e && b != '[' && b != 'O' {
			return
		}
	}
}

// readLineCooked is the fallback for non-terminal stdin (piped input): read a
// line normally, no raw-mode key handling.
func (e *LineEditor) readLineCooked() (string, error) {
	e.write(e.prompt)
	s, err := e.in.ReadString('\n')
	if err != nil {
		if err == io.EOF && s != "" {
			return strings.TrimRight(s, "\r\n"), nil // last line, no newline
		}
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}

func (e *LineEditor) write(s string) { io.WriteString(e.out, s) }
func (e *LineEditor) bell()          { e.write("\a") }
