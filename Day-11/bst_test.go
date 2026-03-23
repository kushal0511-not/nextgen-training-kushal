package main

import (
	"reflect"
	"testing"
)

func TestBSTRelevance(t *testing.T) {
	bst := NewBST()

	// Suggestions for "appl"
	// 1. apple (ED: 1, Freq: 100)
	// 2. apply (ED: 1, Freq: 50)
	// 3. apples (ED: 2, Freq: 200)
	// 4. application (ED: 7, Freq: 500)
	// 5. apricot (ED: 5, Freq: 10)

	// Expectations:
	// - apple (ED: 1, Freq: 100)
	// - apply (ED: 1, Freq: 50)
	// - apples (ED: 2, Freq: 200)
	// - apricot (ED: 5, Freq: 10)
	// - application (ED: 7, Freq: 500)

	bst.Insert("application", 7, 500)
	bst.Insert("apples", 2, 200)
	bst.Insert("apply", 1, 50)
	bst.Insert("apple", 1, 100)
	bst.Insert("apricot", 5, 10)

	expected := []string{"apple", "apply", "apples", "apricot", "application"}
	actual := bst.GetSuggestions()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}

func TestBSTTieBreaker(t *testing.T) {
	bst := NewBST()

	// Same ED and Freq
	bst.Insert("banana", 1, 10)
	bst.Insert("apple", 1, 10)

	expected := []string{"apple", "banana"}
	actual := bst.GetSuggestions()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}
