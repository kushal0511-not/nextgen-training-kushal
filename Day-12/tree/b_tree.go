package tree

type BTreeNode struct {
	Keys     []int
	Children []*BTreeNode
	Leaf     bool
}

type BTree struct {
	Root *BTreeNode
	T    int
}

// NewBTree creates a new B-Tree with the given minimum degree.
func NewBTree(t int) *BTree {
	return &BTree{
		Root: &BTreeNode{
			Keys:     make([]int, 0),
			Children: make([]*BTreeNode, 0),
			Leaf:     true,
		},
		T: t,
	}
}

// Operations: Insert, Search, Delete, RangeQuery(min, max)
// In-order traversal for sorted listing

func (t *BTree) Insert(key int) {
	root := t.Root
	if len(root.Keys) == 2*t.T-1 {
		newRoot := &BTreeNode{
			Keys:     make([]int, 0),
			Children: make([]*BTreeNode, 0),
			Leaf:     false,
		}
		newRoot.Children = append(newRoot.Children, root)
		t.splitChild(newRoot, 0)
		t.Root = newRoot
	}
	t.insertNonFull(t.Root, key)
}

func (t *BTree) splitChild(parent *BTreeNode, i int) {
	child := parent.Children[i]
	newChild := &BTreeNode{
		Keys:     make([]int, 0),
		Children: make([]*BTreeNode, 0),
		Leaf:     child.Leaf,
	}
	medianIndex := t.T - 1
	medianKey := child.Keys[medianIndex]
	newChild.Keys = append(newChild.Keys, child.Keys[medianIndex+1:]...)
	if !child.Leaf {
		newChild.Children = append(newChild.Children, child.Children[t.T:]...)
	}
	child.Keys = child.Keys[:medianIndex]
	child.Children = child.Children[:t.T]
	parent.Keys = append(parent.Keys[:i], append([]int{medianKey}, parent.Keys[i:]...)...)
	parent.Children = append(parent.Children[:i+1], append([]*BTreeNode{newChild}, parent.Children[i+1:]...)...)
}

func (t *BTree) insertNonFull(node *BTreeNode, key int) {
	i := len(node.Keys) - 1
	if node.Leaf {
		node.Keys = append(node.Keys, 0)
		for i >= 0 && key < node.Keys[i] {
			node.Keys[i+1] = node.Keys[i]
			i--
		}
		node.Keys[i+1] = key
	} else {
		for i >= 0 && key < node.Keys[i] {
			i--
		}
		i++
		if len(node.Children[i].Keys) == 2*t.T-1 {
			t.splitChild(node, i)
			if key > node.Keys[i] {
				i++
			}
		}
		t.insertNonFull(node.Children[i], key)
	}
}

func (t *BTree) Search(key int) bool {
	return t.search(t.Root, key)
}

func (t *BTree) search(node *BTreeNode, key int) bool {
	i := 0
	for i < len(node.Keys) && key > node.Keys[i] {
		i++
	}
	if i < len(node.Keys) && key == node.Keys[i] {
		return true
	}
	if node.Leaf {
		return false
	}
	return t.search(node.Children[i], key)
}

func (t *BTree) Delete(key int) {
	t.delete(t.Root, key)
}

func (t *BTree) delete(node *BTreeNode, key int) {
	i := 0
	for i < len(node.Keys) && key > node.Keys[i] {
		i++
	}
	if i < len(node.Keys) && key == node.Keys[i] {
		if node.Leaf {
			t.deleteFromLeaf(node, i)
		} else {
			t.deleteFromNonLeaf(node, i)
		}
	} else {
		if node.Leaf {
			return
		}
		t.delete(node.Children[i], key)
	}
}

func (t *BTree) deleteFromLeaf(node *BTreeNode, i int) {
	node.Keys = append(node.Keys[:i], node.Keys[i+1:]...)
}

func (t *BTree) deleteFromNonLeaf(node *BTreeNode, i int) {
	k := node.Keys[i]
	child := node.Children[i]
	if len(child.Keys) >= t.T {
		predecessor := t.getPredecessor(child)
		node.Keys[i] = predecessor
		t.delete(child, predecessor)
	} else {
		child := node.Children[i+1]
		if len(child.Keys) >= t.T {
			successor := t.getSuccessor(child)
			node.Keys[i] = successor
			t.delete(child, successor)
		} else {
			t.mergeChildren(node, i)
			t.delete(node.Children[i], k)
		}
	}
}

func (t *BTree) getPredecessor(node *BTreeNode) int {
	for len(node.Children) > 0 {
		node = node.Children[len(node.Children)-1]
	}
	return node.Keys[len(node.Keys)-1]
}

func (t *BTree) getSuccessor(node *BTreeNode) int {
	for len(node.Children) > 0 {
		node = node.Children[0]
	}
	return node.Keys[0]
}

func (t *BTree) mergeChildren(node *BTreeNode, i int) {
	child := node.Children[i]
	child.Keys = append(child.Keys, node.Keys[i])
	child.Keys = append(child.Keys, node.Children[i+1].Keys...)
	if !child.Leaf {
		child.Children = append(child.Children, node.Children[i+1].Children...)
	}
	node.Keys = append(node.Keys[:i], node.Keys[i+1:]...)
	node.Children = append(node.Children[:i+1], node.Children[i+2:]...)
}

func (t *BTree) RangeQuery(min, max int) []int {
	result := make([]int, 0)
	t.rangeQuery(t.Root, min, max, &result)
	return result
}

func (t *BTree) rangeQuery(node *BTreeNode, min, max int, result *[]int) {
	i := 0
	for i < len(node.Keys) {
		if !node.Leaf && min < node.Keys[i] {
			t.rangeQuery(node.Children[i], min, max, result)
		}
		if node.Keys[i] >= min && node.Keys[i] <= max {
			*result = append(*result, node.Keys[i])
		}
		if node.Keys[i] > max {
			return
		}
		i++
	}
	if !node.Leaf {
		t.rangeQuery(node.Children[i], min, max, result)
	}
}

func (t *BTree) InOrderTraversal() []int {
	result := make([]int, 0)
	t.inOrderTraversal(t.Root, &result)
	return result
}

func (t *BTree) inOrderTraversal(node *BTreeNode, result *[]int) {
	i := 0
	for i < len(node.Keys) {
		if !node.Leaf {
			t.inOrderTraversal(node.Children[i], result)
		}
		*result = append(*result, node.Keys[i])
		i++
	}
	if !node.Leaf {
		t.inOrderTraversal(node.Children[i], result)
	}
}

// - Visualize tree structure (for debugging)
func (t *BTree) Visualize() {
	printNode(t.Root, 0)
}

func printNode(node *BTreeNode, level int) {
	for i := 0; i < level; i++ {
		print("  ")
	}
	print(node.Keys)
	print("\n")
	if !node.Leaf {
		for _, child := range node.Children {
			printNode(child, level+1)
		}
	}
}
