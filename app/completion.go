package main

import (
	"sort"
	"strings"
)

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

// completions returns every command name starting with prefix, sorted.
func completions(prefix string) []string {
	matches := autoCompleteTrie.wordsWithPrefix(prefix)
	sort.Strings(matches)
	return matches
}

// longestCommonPrefix returns the longest string that is a prefix of every
// input. Command names are ASCII, so byte-wise comparison is fine.
func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}
