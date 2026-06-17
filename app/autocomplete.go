package main

import (
	"os/exec"
	"strings"
)

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
	parts := strings.Fields(string(line))
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

var completers = make(map[string]string)

func (e *LineEditor) autocompleteCompleter(line []rune, scriptPath string, tabCount int) ([]rune, int) {
	parts := strings.Fields(string(line))
	argLength := len(parts) - 1
	var curr, prev string
	if argLength == 0 {
		curr = ""
		prev = ""
	} else if argLength > 0 {
		curr = parts[argLength]
		prev = parts[argLength-1]
	}

	cmd := exec.Command(scriptPath, parts[0], curr, prev)
	output, err := cmd.Output()
	if err != nil {
		e.bell()
		return line, 0
	}

	// One candidate per line. We keep only candidates that have the current
	// word as a prefix: bash trusts the script's output verbatim, but this is a
	// deliberate defensive simplification that guards completed[len(current):]
	// below against a script that ignores argv[2] and dumps a fixed list. (The
	// trade-off: a case-insensitive or fuzzy completer's matches are dropped.)
	var matches []string
	for c := range strings.SplitSeq(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n") {
		if len(c) != 0 && strings.HasPrefix(c, curr) {
			matches = append(matches, c)
		}
	}

	switch len(matches) {
	case 0:
		e.bell()
		return line, 0

	case 1:
		completed := matches[0] + " " // unique match → trailing space for args
		e.write(completed[len(curr):])
		parts[argLength] = completed
		return []rune(strings.Join(parts, " ")), 0

	default:
		// All matches share at least `current`; extend to their common prefix.
		if lcp := longestCommonPrefix(matches); len(lcp) > len(curr) {
			e.write(lcp[len(curr):])
			parts[argLength] = lcp
			return []rune(strings.Join(parts, " ")), 0
		}

		// No further common prefix: bell once, list on the next tab.
		if tabCount == 0 {
			e.bell()
			return line, 1
		}
		e.write("\r\n" + strings.Join(matches, "  ") + "\r\n")
		e.write(e.prompt + string(line))
		return line, 0
	}
}
