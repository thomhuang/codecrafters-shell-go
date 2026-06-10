package main

import (
	"os"
	"sort"
	"strings"
)

// cmdTrie holds every completable command name (builtins plus PATH
// executables), built once at startup.
var cmdTrie *Trie

// cwdFileTrie holds the names of the files in the current directory,
// rebuilt by cdCommand whenever the working directory changes.
var cwdFileTrie *Trie

// same as above but for directories
var cwdDirTrie *Trie

func buildCmdCompletionTrie() {
	t := newTrie()

	for name := range builtins {
		t.insert(name)
	}
	for name := range pathExecutables {
		t.insert(name)
	}

	cmdTrie = t
}

func buildCwdTries() {
	fileTrie, dirTrie := newTrie(), newTrie()

	currPath, _ := os.Getwd()
	currDir, _ := os.ReadDir(currPath)

	for _, entity := range currDir {
		if entity.IsDir() {
			dirTrie.insert(entity.Name())
			continue
		}

		fileTrie.insert(entity.Name())
	}

	cwdFileTrie, cwdDirTrie = fileTrie, dirTrie
}

// getCmdMatches returns every command name starting with prefix, sorted.
func (t *Trie) getCmdMatches(prefix string) []string {
	matches := t.wordsWithPrefix(prefix)
	sort.Strings(matches)
	return matches
}

// getMatches returns every cwd entry starting with prefix, sorted. It reads
// the cached cwdAutoCompleteTrie (built at startup, rebuilt by cdCommand) rather
// than re-scanning the directory on every Tab.
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
