package main

import (
	"os"
	"sort"
	"strings"
)

// cmdTrie holds every completable command name (builtins plus PATH
// executables), built once at startup.
var cmdTrie *Trie

func buildCmdCompletionTrie() {
	t := newTrie()

	for name := range builtins {
		t.insert(name)
	}
	for name := range executables {
		t.insert(name)
	}

	cmdTrie = t
}

// getMatches returns every word in the trie that starts with prefix, sorted. It
// reads the cached trie (built at startup; the cwd tries are rebuilt by
// cdCommand) rather than re-scanning on every Tab.
func (t *Trie) getMatches(prefix string) []string {
	matches := t.wordsWithPrefix(prefix)
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

// isDir reports whether path exists and is a directory, following symlinks.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// getPathMatches returns every entry in dir that starts with prefix.
// If dirsOnly is true, only directories are returned.
func getPathMatches(dir, prefix string, dirsOnly bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if dirsOnly && !entry.IsDir() {
			continue
		}
		matches = append(matches, name)
	}
	sort.Strings(matches)
	return matches
}
