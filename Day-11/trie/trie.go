package trie

type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
	freq        int
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children:    make(map[rune]*TrieNode),
		isEndOfWord: false,
	}
}

func (t *TrieNode) Insert(word string) error {
	//insert word in trie if it exist increment its frequency
	current := t
	for _, char := range word {
		if _, ok := current.children[char]; !ok {
			current.children[char] = NewTrieNode()
		}
		current = current.children[char]
	}
	current.isEndOfWord = true
	current.freq++
	return nil
}

func (t *TrieNode) Search(word string) (bool, int) {
	current := t
	for _, char := range word {
		if _, ok := current.children[char]; !ok {
			return false, 0
		}
		current = current.children[char]
	}
	return current.isEndOfWord, current.freq
}

func (t *TrieNode) StartsWith(prefix string) bool {
	current := t
	for _, char := range prefix {
		if _, ok := current.children[char]; !ok {
			return false
		}
		current = current.children[char]
	}
	return true
}

func (t *TrieNode) AutoComplete(prefix string, limit int) ([]string, []int) {
	current := t
	matches := make([]string, 0)
	freqs := make([]int, 0)
	for _, char := range prefix {
		if _, ok := current.children[char]; !ok {
			return matches, freqs
		}
		current = current.children[char]
	}

	t.isThereEndOfWord(prefix, current, &matches, &freqs, limit)
	return matches, freqs
}

func (t *TrieNode) isThereEndOfWord(stringSoFar string, current *TrieNode, matches *[]string, freqs *[]int, limit int) {
	if len(*matches) >= limit {
		return
	}
	if current.isEndOfWord {
		*matches = append(*matches, stringSoFar)
		*freqs = append(*freqs, current.freq)
	}

	// Recurse to children
	for v, child := range current.children {
		if len(*matches) >= limit {
			return
		}
		t.isThereEndOfWord(stringSoFar+string(v), child, matches, freqs, limit)
	}
}
