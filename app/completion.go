package main

import (
	"os"
	"sort"
	"strings"
)

// cmdAutoCompleteTrie holds every completable command name (builtins plus PATH
// executables), built once at startup.
var cmdAutoCompleteTrie *Trie

// cwdAutoCompleteTrie holds the names of the files in the current directory,
// rebuilt by cdCommand whenever the working directory changes.
var cwdAutoCompleteTrie *Trie

func buildCmdCompletionTrie() *Trie {
	t := newTrie()

	for name := range builtins {
		t.insert(name)
	}
	for name := range pathExecutables {
		t.insert(name)
	}
	return t
}

func buildCwdCompletionTrie() *Trie {
	t := newTrie()

	currPath, _ := os.Getwd()
	currDir, _ := os.ReadDir(currPath)

	for _, file := range currDir {
		if file.IsDir() {
			t.insert(file.Name() + "/")
			continue
		}

		t.insert(file.Name())
	}

	return t
}

// getCmdMatches returns every command name starting with prefix, sorted.
func getCmdMatches(prefix string) []string {
	matches := cmdAutoCompleteTrie.wordsWithPrefix(prefix)
	sort.Strings(matches)
	return matches
}

// getCwdMatches returns every cwd entry starting with prefix, sorted. It reads
// the cached cwdAutoCompleteTrie (built at startup, rebuilt by cdCommand) rather
// than re-scanning the directory on every Tab.
func getCwdMatches(prefix string) []string {
	matches := cwdAutoCompleteTrie.wordsWithPrefix(prefix)
	sort.Strings(matches)
	return matches
}

// For a given provided path and (potentially) partial prefix,
// getPathMatches returns every entry in the directory at path that starts with prefix, sorted. It is used for autocompleting paths like "foo/b" → "foo/bar"
func getPathMatches(path, prefix string) []string {
	t := newTrie()

	currDir, err := os.ReadDir("./" + path)
	if err != nil {
		return []string{}
	}

	for _, file := range currDir {
		if file.IsDir() {
			continue
		}

		t.insert(file.Name())
	}

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
