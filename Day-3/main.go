package main

import (
	"fmt"
	"time"
)

type EditOperation[T any] interface {
	Apply(o Operation[T])
	//This is Reverse
	Undo()
}

type OperationType string

const (
	Add     OperationType = "ADD"
	Delete  OperationType = "DELETE"
	Replace OperationType = "REPLACE"
	//TODO: Fomat
)

type Node[T any] struct {
	Data T
	Next *Node[T]
	Prev *Node[T]
}

func NewNode[T any](data T) *Node[T] {
	return &Node[T]{
		Data: data,
		Next: nil,
		Prev: nil,
	}
}

type DoublyLinkedList[T any] struct {
	Head *Node[T]
	Tail *Node[T]
}
type Operation[T any] struct {
	Position int
	Type     OperationType
	Data     T
	Time     time.Time
	// used in replace and delete

	ExistingText T
}

type Document[T any] struct {
	Doc     DoublyLinkedList[T]
	History []Operation[T]
	// show the version uses idx of history
	CurrentVersion int
}

func NewDocument[T any]() *Document[T] {
	// initialize Doc
	doc := DoublyLinkedList[T]{
		Head: nil,
		Tail: nil,
	}

	// initialize History
	history := make([]Operation[T], 0)

	//return
	return &Document[T]{
		Doc:            doc,
		History:        history,
		CurrentVersion: -1,
	}
}
func (d *Document[T]) Apply(o Operation[T]) {
	// Truncate future history if we're in an undo state
	if d.CurrentVersion < len(d.History)-1 {
		d.History = d.History[:d.CurrentVersion+1]
	}

	switch o.Type {
	case Add:
		d.applyAdd(o.Position, o.Data)
	case Delete:
		// Capture ExistingText from the node that will be deleted
		if curr := d.findNode(o.Position); curr != nil {
			o.ExistingText = curr.Data
		}
		d.applyDelete(o.Position)
	case Replace:
		// Capture ExistingText from the node that will be replaced
		if curr := d.findNode(o.Position); curr != nil {
			o.ExistingText = curr.Data
		}
		d.applyReplace(o.Position, o.Data)
	}
	d.History = append(d.History, o)
	d.CurrentVersion++
}

func (d *Document[T]) findNode(pos int) *Node[T] {
	if d.Doc.Head == nil {
		return nil
	}
	curr := d.Doc.Head
	for i := 0; i < pos && curr.Next != nil; i++ {
		curr = curr.Next
	}
	return curr
}

func (d *Document[T]) Undo() {
	if d.CurrentVersion < 0 {
		return
	}

	op := d.History[d.CurrentVersion]
	switch op.Type {
	case Add:
		// To undo an add, we delete at that position
		d.applyDelete(op.Position)
	case Delete:
		// To undo a delete, we add the existing text back
		d.applyAdd(op.Position, op.ExistingText)
	case Replace:
		// To undo a replace, we put back the existing text
		d.applyReplace(op.Position, op.ExistingText)
	}
	d.CurrentVersion--
}

func (d *Document[T]) Redo() {
	if d.CurrentVersion >= len(d.History)-1 {
		return
	}

	d.CurrentVersion++
	op := d.History[d.CurrentVersion]
	switch op.Type {
	case Add:
		d.applyAdd(op.Position, op.Data)
	case Delete:
		d.applyDelete(op.Position)
	case Replace:
		d.applyReplace(op.Position, op.Data)
	}
}

// Helper methods to avoid code duplication and handle list logic cleanly
func (d *Document[T]) applyAdd(pos int, data T) {
	node := NewNode(data)
	if d.Doc.Head == nil {
		d.Doc.Head = node
		d.Doc.Tail = node
		return
	}

	if pos == 0 {
		node.Next = d.Doc.Head
		d.Doc.Head.Prev = node
		d.Doc.Head = node
		return
	}

	// Insert after the node at rank pos-1 (or the tail)
	curr := d.Doc.Head
	for i := 0; i < pos-1 && curr.Next != nil; i++ {
		curr = curr.Next
	}

	node.Next = curr.Next
	node.Prev = curr
	if curr.Next != nil {
		curr.Next.Prev = node
	} else {
		d.Doc.Tail = node
	}
	curr.Next = node
}

func (d *Document[T]) applyDelete(pos int) {
	if d.Doc.Head == nil {
		return
	}

	curr := d.Doc.Head
	if pos == 0 {
		d.Doc.Head = curr.Next
		if d.Doc.Head != nil {
			d.Doc.Head.Prev = nil
		} else {
			d.Doc.Tail = nil
		}
		return
	}

	// Delete the node at rank pos (or the tail)
	for i := 0; i < pos && curr.Next != nil; i++ {
		curr = curr.Next
	}

	if curr != nil {
		if curr.Prev != nil {
			curr.Prev.Next = curr.Next
		}
		if curr.Next != nil {
			curr.Next.Prev = curr.Prev
		} else {
			d.Doc.Tail = curr.Prev
		}
	}
}

func (d *Document[T]) applyReplace(pos int, data T) {
	curr := d.findNode(pos)
	if curr != nil {
		curr.Data = data
	}
}
func (dll DoublyLinkedList[T]) String() string {
	var result string
	curr := dll.Head
	for curr != nil {
		result += fmt.Sprint(curr.Data)
		curr = curr.Next
	}
	return result
}

func (d *Document[T]) String() string {
	return d.Doc.String()
}

func main() {
	doc := NewDocument[string]()

	// 1. insert 0 "Hello" -> doc: "Hello"
	fmt.Println("--- Step 1: Insert 0 'Hello' ---")
	doc.Apply(Operation[string]{Position: 0, Type: Add, Data: "Hello", Time: time.Now()})
	fmt.Printf("doc: \"%s\"\n", doc)

	// 2. insert 5 " World" -> doc: "Hello World"
	fmt.Println("\n--- Step 2: Insert 5 ' World' ---")
	doc.Apply(Operation[string]{Position: 5, Type: Add, Data: " World", Time: time.Now()})
	fmt.Printf("doc: \"%s\"\n", doc)

	// 3. undo -> doc: "Hello"
	fmt.Println("\n--- Step 3: Undo ---")
	doc.Undo()
	fmt.Printf("doc: \"%s\"\n", doc)

	// 4. undo -> doc: ""
	fmt.Println("\n--- Step 4: Undo ---")
	doc.Undo()
	fmt.Printf("doc: \"%s\"\n", doc)

	// 5. redo -> doc: "Hello"
	fmt.Println("\n--- Step 5: Redo ---")
	doc.Redo()
	fmt.Printf("doc: \"%s\"\n", doc)

	// 6. insert 5 " Go!" -> doc: "Hello Go!"
	fmt.Println("\n--- Step 6: Insert 5 ' Go!' ---")
	doc.Apply(Operation[string]{Position: 5, Type: Add, Data: " Go!", Time: time.Now()})
	fmt.Printf("doc: \"%s\"\n", doc)
}
