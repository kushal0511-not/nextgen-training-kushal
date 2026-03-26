package category

import (
	"fmt"
	"strings"
	"sync"
)

// CategoryNode represents a node in the category tree.
type CategoryNode struct {
	Name          string
	SubCategories map[string]*CategoryNode
}

// NewCategoryNode creates a new CategoryNode.
func NewCategoryNode(name string) *CategoryNode {
	return &CategoryNode{
		Name:          name,
		SubCategories: make(map[string]*CategoryNode),
	}
}

// AddSubCategory adds a subcategory to the current node.
func (n *CategoryNode) AddSubCategory(name string) *CategoryNode {
	if _, exists := n.SubCategories[name]; !exists {
		n.SubCategories[name] = NewCategoryNode(name)
	}
	return n.SubCategories[name]
}

// FindCategory finds a category by its path (e.g., ["Electronics", "Phones", "Smartphones"]).
func (n *CategoryNode) FindCategory(path []string) (*CategoryNode, error) {
	if len(path) == 0 {
		return n, nil
	}
	subName := path[0]
	sub, exists := n.SubCategories[subName]
	if !exists {
		return nil, fmt.Errorf("category %s not found", strings.Join(path, " > "))
	}
	return sub.FindCategory(path[1:])
}

// FindByName searches for a category by name anywhere in the subtree.
func (n *CategoryNode) FindByName(name string) *CategoryNode {
	if n.Name == name {
		return n
	}
	for _, sub := range n.SubCategories {
		if found := sub.FindByName(name); found != nil {
			return found
		}
	}
	return nil
}

// GetAllCategoryNames returns the names of the category and all its subcategories.
func (n *CategoryNode) GetAllCategoryNames() []string {
	names := []string{n.Name}
	for _, sub := range n.SubCategories {
		names = append(names, sub.GetAllCategoryNames()...)
	}
	return names
}

// PrintTree prints the category tree for debugging.
func (n *CategoryNode) PrintTree(indent string) {
	fmt.Printf("%s%s\n", indent, n.Name)
	for _, sub := range n.SubCategories {
		sub.PrintTree(indent + "  ")
	}
}

// CategoryTree managed the root of the category structure with concurrency safety.
type CategoryTree struct {
	Root *CategoryNode
	mu   sync.RWMutex
}

// AddCategory adds a category path to the tree safely.
func (t *CategoryTree) AddCategory(path []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	current := t.Root
	for _, name := range path {
		current = current.AddSubCategory(name)
	}
}

// FindCategory finds a category by its path safely.
func (t *CategoryTree) FindCategory(path []string) (*CategoryNode, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// Try path-based search first.
	node, err := t.Root.FindCategory(path)
	if err == nil {
		return node, nil
	}
	// If it's a single name, try global search.
	if len(path) == 1 {
		if node := t.Root.FindByName(path[0]); node != nil {
			return node, nil
		}
	}
	return nil, err
}

// PrintTree prints the category tree safely.
func (t *CategoryTree) PrintTree() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.Root.PrintTree("")
}

// NewCategoryTree initializes a CategoryTree with a root node (usually "All").
func NewCategoryTree() *CategoryTree {
	return &CategoryTree{
		Root: NewCategoryNode("All"),
	}
}
