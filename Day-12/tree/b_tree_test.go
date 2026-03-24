package tree

import "testing"

// Unit tests for B-Tree operations

func TestBTreeInsertAndSearch(t *testing.T) {
	tree := NewBTree(8)
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(5)
	tree.Insert(6)
	tree.Insert(12)
	tree.Insert(30)
	tree.Insert(7)
	tree.Insert(17)
	found := tree.Search(5)
	if !found {
		t.Errorf("Expected 5 to be found")
	}
	found = tree.Search(100)
	if found {
		t.Errorf("Expected 100 to not be found")
	}
}

func TestBTreeDelete(t *testing.T) {
	tree := NewBTree(8)
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(5)
	tree.Insert(6)
	tree.Insert(12)
	tree.Insert(30)
	tree.Insert(7)
	tree.Insert(17)
	tree.Delete(5)
	found := tree.Search(5)
	if found {
		t.Errorf("Expected 5 to not be found")
	}
}

func TestBTreeRangeQuery(t *testing.T) {
	tree := NewBTree(8)
	tree.Insert(10)
	tree.Insert(20)
	tree.Insert(5)
	tree.Insert(6)
	tree.Insert(12)
	tree.Insert(30)
	tree.Insert(7)
	tree.Insert(17)
	results := tree.RangeQuery(6, 50)
	if len(results) != 7 {
		t.Errorf("Expected 7 results, got %d", len(results))
	}
}
