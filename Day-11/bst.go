package main

type BST struct {
	root *BSTNode
}

type BSTNode struct {
	Data         string
	EditDistance int
	Freq         int
	Left         *BSTNode
	Right        *BSTNode
}

func NewBST() *BST {
	return &BST{}
}

// Insert adds a new suggestion to the BST based on relevance.
// Relevance is determined by:
// 1. Lower EditDistance is better (comes first in in-order traversal).
// 2. Higher Freq is better (comes first for the same EditDistance).
// 3. Alphabetical order as a tie-breaker.
func (b *BST) Insert(data string, editDistance int, freq int) {
	newNode := &BSTNode{
		Data:         data,
		EditDistance: editDistance,
		Freq:         freq,
	}

	if b.root == nil {
		b.root = newNode
		return
	}

	current := b.root
	for {
		if b.isBetter(newNode, current) {
			if current.Left == nil {
				current.Left = newNode
				return
			}
			current = current.Left
		} else if b.isBetter(current, newNode) {
			if current.Right == nil {
				current.Right = newNode
				return
			}
			current = current.Right
		} else {
			// Duplicate (same Data, EditDistance, and Freq) - ignore or update
			return
		}
	}
}

// isBetter returns true if node A is "better" than node B.
// "Better" means it should appear earlier in an in-order traversal.
func (b *BST) isBetter(a, bNode *BSTNode) bool {
	if a.EditDistance < bNode.EditDistance {
		return true
	}
	if a.EditDistance > bNode.EditDistance {
		return false
	}
	// Edit distances are equal, check frequency (higher is better)
	if a.Freq > bNode.Freq {
		return true
	}
	if a.Freq < bNode.Freq {
		return false
	}
	// Frequencies are equal, check alphabetical order
	return a.Data < bNode.Data
}

// GetSuggestions returns suggestions in order of relevance using in-order traversal.
func (b *BST) GetSuggestions() []string {
	var suggestions []string
	b.inOrder(b.root, &suggestions)
	return suggestions
}

func (b *BST) inOrder(node *BSTNode, suggestions *[]string) {
	if node == nil {
		return
	}
	b.inOrder(node.Left, suggestions)
	*suggestions = append(*suggestions, node.Data)
	b.inOrder(node.Right, suggestions)
}

func (b *BST) Clear() {
	b.root = nil
}
