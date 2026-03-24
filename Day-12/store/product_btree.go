package store

type Item struct {
	Price float64
	IDs   []string
}

type ProductBTreeNode struct {
	Items    []Item
	Children []*ProductBTreeNode
	Leaf     bool
}

type ProductPriceBTree struct {
	Root *ProductBTreeNode
	T    int
}

func NewProductPriceBTree(t int) *ProductPriceBTree {
	return &ProductPriceBTree{
		Root: &ProductBTreeNode{
			Items:    make([]Item, 0),
			Children: make([]*ProductBTreeNode, 0),
			Leaf:     true,
		},
		T: t,
	}
}

func (t *ProductPriceBTree) Insert(price float64, id string) {
	// If price already exists, just append ID
	node, index := t.searchNode(t.Root, price)
	if node != nil {
		node.Items[index].IDs = append(node.Items[index].IDs, id)
		return
	}

	root := t.Root
	if len(root.Items) == 2*t.T-1 {
		newRoot := &ProductBTreeNode{
			Items:    make([]Item, 0),
			Children: make([]*ProductBTreeNode, 0),
			Leaf:     false,
		}
		newRoot.Children = append(newRoot.Children, root)
		t.splitChild(newRoot, 0)
		t.Root = newRoot
	}
	t.insertNonFull(t.Root, price, id)
}

func (t *ProductPriceBTree) searchNode(node *ProductBTreeNode, price float64) (*ProductBTreeNode, int) {
	i := 0
	for i < len(node.Items) && price > node.Items[i].Price {
		i++
	}
	if i < len(node.Items) && price == node.Items[i].Price {
		return node, i
	}
	if node.Leaf {
		return nil, -1
	}
	return t.searchNode(node.Children[i], price)
}

func (t *ProductPriceBTree) splitChild(parent *ProductBTreeNode, i int) {
	child := parent.Children[i]
	newChild := &ProductBTreeNode{
		Items:    make([]Item, 0),
		Children: make([]*ProductBTreeNode, 0),
		Leaf:     child.Leaf,
	}
	medianIndex := t.T - 1
	medianItem := child.Items[medianIndex]

	newChild.Items = append(newChild.Items, child.Items[medianIndex+1:]...)
	if !child.Leaf {
		newChild.Children = append(newChild.Children, child.Children[t.T:]...)
	}
	child.Items = child.Items[:medianIndex]
	if !child.Leaf {
		child.Children = child.Children[:t.T]
	}

	parent.Items = append(parent.Items[:i], append([]Item{medianItem}, parent.Items[i:]...)...)
	parent.Children = append(parent.Children[:i+1], append([]*ProductBTreeNode{newChild}, parent.Children[i+1:]...)...)
}

func (t *ProductPriceBTree) insertNonFull(node *ProductBTreeNode, price float64, id string) {
	i := len(node.Items) - 1
	if node.Leaf {
		node.Items = append(node.Items, Item{})
		for i >= 0 && price < node.Items[i].Price {
			node.Items[i+1] = node.Items[i]
			i--
		}
		node.Items[i+1] = Item{Price: price, IDs: []string{id}}
	} else {
		for i >= 0 && price < node.Items[i].Price {
			i--
		}
		i++
		if len(node.Children[i].Items) == 2*t.T-1 {
			t.splitChild(node, i)
			if price > node.Items[i].Price {
				i++
			}
		}
		t.insertNonFull(node.Children[i], price, id)
	}
}

func (t *ProductPriceBTree) Delete(price float64, id string) {
	t.deleteItem(t.Root, price, id)
}

func (t *ProductPriceBTree) deleteItem(node *ProductBTreeNode, price float64, id string) {
	i := 0
	for i < len(node.Items) && price > node.Items[i].Price {
		i++
	}

	if i < len(node.Items) && price == node.Items[i].Price {
		// remove ID from IDs
		ids := node.Items[i].IDs
		for j, v := range ids {
			if v == id {
				node.Items[i].IDs = append(ids[:j], ids[j+1:]...)
				break
			}
		}
		if len(node.Items[i].IDs) > 0 {
			// still has IDs, so we don't remove the key from BTree
			return
		}

		if node.Leaf {
			t.deleteFromLeaf(node, i)
		} else {
			t.deleteFromNonLeaf(node, i)
		}
	} else {
		if node.Leaf {
			return
		}
		t.deleteItem(node.Children[i], price, id)
	}
}

func (t *ProductPriceBTree) deleteFromLeaf(node *ProductBTreeNode, i int) {
	node.Items = append(node.Items[:i], node.Items[i+1:]...)
}

func (t *ProductPriceBTree) deleteFromNonLeaf(node *ProductBTreeNode, i int) {
	k := node.Items[i]
	child := node.Children[i]
	if len(child.Items) >= t.T {
		predecessor := t.getPredecessor(child)
		node.Items[i] = predecessor
		// Recursively delete the predecessor key (we pass its first ID temporarily since deleteItem will clear all anyway but we want to only delete the node key)
		t.deleteItemFullNode(child, predecessor.Price)
	} else {
		child2 := node.Children[i+1]
		if len(child2.Items) >= t.T {
			successor := t.getSuccessor(child2)
			node.Items[i] = successor
			t.deleteItemFullNode(child2, successor.Price)
		} else {
			t.mergeChildren(node, i)
			t.deleteItemFullNode(node.Children[i], k.Price)
		}
	}
}

// deleteItemFullNode is a helper to forcefully delete an entire Item by price
func (t *ProductPriceBTree) deleteItemFullNode(node *ProductBTreeNode, price float64) {
	i := 0
	for i < len(node.Items) && price > node.Items[i].Price {
		i++
	}
	if i < len(node.Items) && price == node.Items[i].Price {
		if node.Leaf {
			t.deleteFromLeaf(node, i)
		} else {
			t.deleteFromNonLeaf(node, i)
		}
	} else if !node.Leaf {
		t.deleteItemFullNode(node.Children[i], price)
	}
}

func (t *ProductPriceBTree) getPredecessor(node *ProductBTreeNode) Item {
	for len(node.Children) > 0 {
		node = node.Children[len(node.Children)-1]
	}
	return node.Items[len(node.Items)-1]
}

func (t *ProductPriceBTree) getSuccessor(node *ProductBTreeNode) Item {
	for len(node.Children) > 0 {
		node = node.Children[0]
	}
	return node.Items[0]
}

func (t *ProductPriceBTree) mergeChildren(node *ProductBTreeNode, i int) {
	child := node.Children[i]
	child.Items = append(child.Items, node.Items[i])
	child.Items = append(child.Items, node.Children[i+1].Items...)
	if !child.Leaf {
		child.Children = append(child.Children, node.Children[i+1].Children...)
	}
	node.Items = append(node.Items[:i], node.Items[i+1:]...)
	node.Children = append(node.Children[:i+1], node.Children[i+2:]...)
}

func (t *ProductPriceBTree) RangeQuery(min, max float64) []string {
	var result []string
	if t.Root != nil {
		t.rangeQuery(t.Root, min, max, &result)
	}
	return result
}

func (t *ProductPriceBTree) rangeQuery(node *ProductBTreeNode, min, max float64, result *[]string) {
	if node == nil {
		return
	}
	i := 0
	for i < len(node.Items) && min > node.Items[i].Price {
		i++
	}
	for i < len(node.Items) && node.Items[i].Price <= max {
		if !node.Leaf {
			t.rangeQuery(node.Children[i], min, max, result)
		}
		if node.Items[i].Price >= min {
			*result = append(*result, node.Items[i].IDs...)
		}
		i++
	}
	if !node.Leaf {
		t.rangeQuery(node.Children[i], min, max, result)
	}
}

func (t *ProductPriceBTree) InOrderTraversal() []string {
	var result []string
	if t.Root != nil {
		t.inOrderTraversal(t.Root, &result)
	}
	return result
}

func (t *ProductPriceBTree) inOrderTraversal(node *ProductBTreeNode, result *[]string) {
	if node == nil {
		return
	}
	i := 0
	for i < len(node.Items) {
		if !node.Leaf {
			t.inOrderTraversal(node.Children[i], result)
		}
		*result = append(*result, node.Items[i].IDs...)
		i++
	}
	if !node.Leaf {
		t.inOrderTraversal(node.Children[i], result)
	}
}

func (t *ProductPriceBTree) InOrderTraversalPaginated(offset, limit int) []string {
	var result []string
	if t.Root != nil {
		count := 0
		t.inOrderTraversalPaginated(t.Root, offset, limit, &result, &count)
	}
	return result
}

func (t *ProductPriceBTree) inOrderTraversalPaginated(node *ProductBTreeNode, offset, limit int, result *[]string, count *int) {
	if node == nil || len(*result) >= limit {
		return
	}
	i := 0
	for i < len(node.Items) {
		if !node.Leaf {
			t.inOrderTraversalPaginated(node.Children[i], offset, limit, result, count)
		}
		if len(*result) >= limit {
			return
		}
		for _, id := range node.Items[i].IDs {
			if *count >= offset && len(*result) < limit {
				*result = append(*result, id)
			}
			*count++
		}
		i++
	}
	if !node.Leaf {
		t.inOrderTraversalPaginated(node.Children[i], offset, limit, result, count)
	}
}
