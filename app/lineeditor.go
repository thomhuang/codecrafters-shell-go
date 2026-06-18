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
	cycle  *dirCycle // tab-cycling for directory autocomplete
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
	e.cycle = nil // no directory cycle carries over from a previous line
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
				e.eraseRunes(1) // erase the last rune
			} else {
				e.bell() // nothing to delete
			}
			tabCount = 0
			e.cycle = nil // editing cancels a directory cycle

		case r == '\t': // tab autocomplete
			cmd, _, hasSpace := strings.Cut(string(line), " ")
			if !hasSpace {
				// Autocomplete for commands (the first word).
				line, tabCount = e.autocompleteCommand(line, tabCount)
			} else if scriptPath, exists := completerScripts[cmd]; cmd != "" && exists {
				// Autocomplete for commands with custom completers.
				line, tabCount = e.autocompleteCompleterScript(line, scriptPath, tabCount)
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
			e.cycle = nil // editing cancels a directory cycle
		}
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

func (e *LineEditor) eraseRunes(n int) {
	e.write(strings.Repeat("\b \b", n))
}
