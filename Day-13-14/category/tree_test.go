package category

import (
	"fmt"
	"sync"
	"testing"
)

func TestCategoryTree(t *testing.T) {
	tree := NewCategoryTree()

	// Test AddCategory (Thread-safe)
	tree.AddCategory([]string{"Electronics", "Phones", "Smartphones"})

	// Test FindCategory (Thread-safe)
	path := []string{"Electronics", "Phones", "Smartphones"}
	node, err := tree.FindCategory(path)
	if err != nil {
		t.Fatalf("Failed to find category: %v", err)
	}
	if node.Name != "Smartphones" {
		t.Errorf("Expected Smartphones, got %s", node.Name)
	}

	// Test non-existent path
	_, err = tree.FindCategory([]string{"Electronics", "Laptops"})
	if err == nil {
		t.Errorf("Expected error for non-existent category")
	}
}

func TestCategoryTreeStress(t *testing.T) {
	tree := NewCategoryTree()
	const numGoroutines = 50
	const depth = 5
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			path := []string{"Root"}
			for d := 0; d < depth; d++ {
				path = append(path, fmt.Sprintf("Level%d-Node%d", d, id))
				tree.AddCategory(path)
				_, _ = tree.FindCategory(path)
			}
		}(i)
	}

	wg.Wait()
}
