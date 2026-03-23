package trie

import "testing"

func TestInsertAndSearch(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("apple")
	trie.Insert("apple")
	trie.Insert("banana")

	found, freq := trie.Search("apple")
	if !found || freq != 2 {
		t.Errorf("Expected apple to be found with frequency 2, got %v, %d", found, freq)
	}

	found, freq = trie.Search("banana")
	if !found || freq != 1 {
		t.Errorf("Expected banana to be found with frequency 1, got %v, %d", found, freq)
	}

	found, freq = trie.Search("orange")
	if found || freq != 0 {
		t.Errorf("Expected orange not to be found, got %v, %d", found, freq)
	}
}

func TestStartsWith(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("apple")
	trie.Insert("banana")

	if !trie.StartsWith("app") {
		t.Errorf("Expected apple to have prefix app")
	}

	if !trie.StartsWith("ban") {
		t.Errorf("Expected banana to have prefix ban")
	}

	if trie.StartsWith("ora") {
		t.Errorf("Expected orange not to have prefix ora")
	}
}

func TestAutoComplete(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("apple")
	trie.Insert("app")
	trie.Insert("apricot")
	trie.Insert("banana")

	matches, freqs := trie.AutoComplete("ap", 2)
	if len(matches) != 2 || len(freqs) != 2 {
		t.Errorf("Expected 2 matches for prefix 'ap', got %d", len(matches))
	}

	// Check if expected words are present (order might vary depending on map iteration, but content should match)
	expected := map[string]bool{"apple": true, "app": true, "apricot": true}
	for _, match := range matches {
		if !expected[match] {
			t.Errorf("Unexpected word in autocomplete: %s", match)
		}
	}
}

func TestAutoCompleteLimit(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("apple")
	trie.Insert("app")
	trie.Insert("apricot")

	matches, _ := trie.AutoComplete("ap", 1)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match due to limit, got %d", len(matches))
	}
}

func TestAutoCompleteNoMatch(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("apple")

	matches, freqs := trie.AutoComplete("ora", 5)
	if len(matches) != 0 || len(freqs) != 0 {
		t.Errorf("Expected no matches for prefix 'ora', got %d matches", len(matches))
	}
}
