package main

import "sort"

// autoCompleteTrie holds every completable command name (builtins plus PATH
// executables), built once at startup.
var autoCompleteTrie *Trie

func buildCompletionTrie() *Trie {
	t := newTrie()
	for name := range builtins {
		t.insert(name)
	}
	for name := range pathExecutables {
		t.insert(name)
	}
	return t
}

type commandAutoComplete struct {
	trie *Trie
}

// "interface" method for readline's AutoCompleter
// This is called with the current input and cursor position, and expects a list of possible suffixes to complete the current word
// along with the length of the already-typed prefix.
// We only autocomplete the first word (the command), so if the cursor is past the first space, we return no completions.
func (c *commandAutoComplete) Do(line []rune, pos int) ([][]rune, int) {
	// Find the start of the word under the cursor.
	start := pos
	for start > 0 && line[start-1] != ' ' {
		start--
	}

	// if start != 0, then we were presented with args ... so we don't do any command autocompletion
	if start != 0 {
		return nil, 0
	}

	// start will be 0! so the prefix is just the first word up to the cursor position
	prefix := string(line[start:pos])
	matches := c.trie.wordsWithPrefix(prefix)
	sort.Strings(matches)

	suffixes := make([][]rune, 0, len(matches))
	for _, m := range matches {
		suffix := m[len(prefix):] // strip the already-typed prefix
		if len(matches) == 1 {
			suffix += " " // exact single completion → trailing space for additional args :P
		}
		suffixes = append(suffixes, []rune(suffix))
	}
	return suffixes, len([]rune(prefix))
}
