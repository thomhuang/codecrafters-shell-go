package main

type Trie struct {
	children    map[string]*Trie
	isEndOfWord bool // marks the end of an inserted word
}

func newTrie() *Trie {
	return &Trie{children: make(map[string]*Trie)}
}

func (t *Trie) insert(word string) {
	node := t
	for _, r := range word {
		head := string(r)
		child, exists := node.children[head]
		if !exists {
			child = newTrie()
			node.children[head] = child
		}
		node = child
	}
	node.isEndOfWord = true
}

// search walks to the node reached by word, or returns nil if no such path
// exists.
func (t *Trie) search(word string) *Trie {
	node := t
	for _, r := range word {
		child, exists := node.children[string(r)]
		if !exists {
			return nil
		}
		node = child
	}
	return node
}

// wordsWithPrefix returns every inserted word that starts with prefix.
func (t *Trie) wordsWithPrefix(prefix string) []string {
	node := t.search(prefix)
	if node == nil {
		return nil
	}

	var out []string
	node.collect(prefix, &out)
	return out
}

// collect appends to out every word reachable from t, where prefix is the
// string accumulated on the way to t.
func (t *Trie) collect(prefix string, out *[]string) {
	if t.isEndOfWord {
		*out = append(*out, prefix)
	}
	for ch, child := range t.children {
		child.collect(prefix+ch, out)
	}
}
