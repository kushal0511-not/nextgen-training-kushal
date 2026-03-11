package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultNodeCap = 10

type OperationType string

const (
	Add     OperationType = "ADD"
	Delete  OperationType = "DELETE"
	Replace OperationType = "REPLACE"
	Format  OperationType = "FORMAT"
)

type NodeMetadata map[string]bool

type Node[T any] struct {
	Content  T
	Cap      int
	StrLen   int
	Metadata NodeMetadata
	Next     *Node[T]
	Prev     *Node[T]
}

func NewNode(content string, cap int) *Node[string] {
	if cap <= 0 {
		cap = DefaultNodeCap
	}
	return &Node[string]{
		Content:  content,
		Cap:      cap,
		StrLen:   len(content),
		Metadata: make(NodeMetadata),
	}
}

type Operation struct {
	Type         OperationType
	Position     int
	Length       int
	Data         string
	Metadata     NodeMetadata
	ExistingText string
	ExistingMeta []NodeMetadata
	Timestamp    time.Time
}

type DoublyLinkedList struct {
	Head *Node[string]
	Tail *Node[string]
}

type Document struct {
	Doc            DoublyLinkedList
	History        []Operation
	CurrentVersion int
}

func NewDocument() *Document {
	return &Document{
		History:        make([]Operation, 0),
		CurrentVersion: -1,
	}
}

func (d *Document) findNodeAtCharPos(pos int) (*Node[string], int) {
	if d.Doc.Head == nil {
		return nil, 0
	}
	curr := d.Doc.Head
	accum := 0
	for curr != nil {
		if pos >= accum && pos < accum+curr.StrLen {
			return curr, pos - accum
		}
		accum += curr.StrLen
		curr = curr.Next
	}
	if pos == accum {
		return d.Doc.Tail, d.Doc.Tail.StrLen
	}
	return nil, 0
}

func (d *Document) Apply(o Operation) {
	if d.CurrentVersion < len(d.History)-1 {
		d.History = d.History[:d.CurrentVersion+1]
	}

	o.Timestamp = time.Now()

	// If inserting or replacing beyond the document end, pad with spaces.
	// We modify the Operation itself so that Undo() knows to remove the padding.
	currentLen := d.docLen()
	if (o.Type == Add || o.Type == Replace) && o.Position > currentLen {
		padding := strings.Repeat(" ", o.Position-currentLen)
		o.Data = padding + o.Data
		o.Position = currentLen
	}

	switch o.Type {
	case Add:
		d.applyAdd(o.Position, o.Data)
	case Delete:
		o.ExistingText = d.applyDelete(o.Position, o.Length)
	case Replace:
		oldText := d.applyDelete(o.Position, o.Length)
		o.ExistingText = oldText
		d.applyAdd(o.Position, o.Data)
	case Format:
		o.ExistingMeta = d.applyFormat(o.Position, o.Length, o.Metadata)
	}

	d.History = append(d.History, o)
	d.CurrentVersion++
}

func (d *Document) docLen() int {
	n := 0
	curr := d.Doc.Head
	for curr != nil {
		n += curr.StrLen
		curr = curr.Next
	}
	return n
}

func (d *Document) applyAdd(pos int, data string) {
	if d.Doc.Head == nil {
		d.insertNewNodes(nil, data)
		return
	}

	node, offset := d.findNodeAtCharPos(pos)
	if node == nil {
		// pos is exactly at the end of the document — append
		d.insertNewNodes(d.Doc.Tail, data)
		return
	}

	before := node.Content[:offset]
	after := node.Content[offset:]
	newContent := before + data + after

	if len(newContent) <= node.Cap {
		node.Content = newContent
		node.StrLen = len(newContent)
	} else {
		d.redistribute(node, newContent)
	}
}

func (d *Document) insertNewNodes(afterNode *Node[string], data string) {
	for len(data) > 0 {
		chunkSize := DefaultNodeCap
		if len(data) < DefaultNodeCap {
			chunkSize = len(data)
		}
		chunk := data[:chunkSize]
		data = data[chunkSize:]

		newNode := NewNode(chunk, DefaultNodeCap)
		if afterNode == nil {
			if d.Doc.Head == nil {
				d.Doc.Head = newNode
				d.Doc.Tail = newNode
			} else {
				newNode.Next = d.Doc.Head
				d.Doc.Head.Prev = newNode
				d.Doc.Head = newNode
			}
		} else {
			newNode.Next = afterNode.Next
			newNode.Prev = afterNode
			if afterNode.Next != nil {
				afterNode.Next.Prev = newNode
			} else {
				d.Doc.Tail = newNode
			}
			afterNode.Next = newNode
		}
		afterNode = newNode
	}
}

func (d *Document) redistribute(node *Node[string], content string) {
	metadata := node.Metadata

	node.Content = content[:node.Cap]
	node.StrLen = node.Cap
	remaining := content[node.Cap:]

	curr := node
	for len(remaining) > 0 {
		chunkSize := node.Cap
		if len(remaining) < node.Cap {
			chunkSize = len(remaining)
		}
		chunk := remaining[:chunkSize]
		remaining = remaining[chunkSize:]

		newNode := NewNode(chunk, node.Cap)
		newNode.Metadata = make(NodeMetadata)
		for k, v := range metadata {
			newNode.Metadata[k] = v
		}

		newNode.Next = curr.Next
		newNode.Prev = curr
		if curr.Next != nil {
			curr.Next.Prev = newNode
		} else {
			d.Doc.Tail = newNode
		}
		curr.Next = newNode
		curr = newNode
	}
}

func (d *Document) applyDelete(pos int, length int) string {
	if d.Doc.Head == nil || length <= 0 {
		return ""
	}

	var deletedText strings.Builder
	remainingLen := length

	for remainingLen > 0 {
		node, offset := d.findNodeAtCharPos(pos)
		if node == nil || (offset == node.StrLen && node.Next == nil) {
			break
		}

		if offset == node.StrLen {
			node = node.Next
			if node == nil {
				break
			}
			offset = 0
		}

		canDelete := node.StrLen - offset
		if canDelete > remainingLen {
			canDelete = remainingLen
		}

		deletedPart := node.Content[offset : offset+canDelete]
		deletedText.WriteString(deletedPart)

		node.Content = node.Content[:offset] + node.Content[offset+canDelete:]
		node.StrLen = len(node.Content)

		if node.StrLen == 0 {
			d.removeNode(node)
		}

		remainingLen -= canDelete
	}

	return deletedText.String()
}

func (d *Document) removeNode(node *Node[string]) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		d.Doc.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		d.Doc.Tail = node.Prev
	}
}

func (d *Document) applyFormat(pos int, length int, meta NodeMetadata) []NodeMetadata {
	if d.Doc.Head == nil || length <= 0 {
		return nil
	}

	var existingMeta []NodeMetadata
	remainingLen := length
	currentPos := pos

	for remainingLen > 0 {
		node, offset := d.findNodeAtCharPos(currentPos)
		if node == nil || (offset == node.StrLen && node.Next == nil) {
			break
		}

		if offset == node.StrLen {
			node = node.Next
			if node == nil {
				break
			}
			offset = 0
		}

		if offset > 0 {
			d.splitNode(node, offset)
			node = node.Next
			offset = 0
		}

		nodeLen := node.StrLen
		if nodeLen > remainingLen {
			d.splitNode(node, remainingLen)
			nodeLen = remainingLen
		}

		oldMeta := make(NodeMetadata)
		for k, v := range node.Metadata {
			oldMeta[k] = v
		}
		existingMeta = append(existingMeta, oldMeta)

		for k, v := range meta {
			node.Metadata[k] = v
		}

		remainingLen -= nodeLen
		currentPos += nodeLen
	}

	return existingMeta
}

func (d *Document) splitNode(node *Node[string], offset int) {
	if offset <= 0 || offset >= node.StrLen {
		return
	}

	newContent := node.Content[offset:]
	node.Content = node.Content[:offset]
	node.StrLen = offset

	newNode := NewNode(newContent, node.Cap)
	newNode.Metadata = make(NodeMetadata)
	for k, v := range node.Metadata {
		newNode.Metadata[k] = v
	}

	newNode.Next = node.Next
	newNode.Prev = node
	if node.Next != nil {
		node.Next.Prev = newNode
	} else {
		d.Doc.Tail = newNode
	}
	node.Next = newNode
}

func (d *Document) Undo() {
	if d.CurrentVersion < 0 {
		return
	}

	op := d.History[d.CurrentVersion]
	switch op.Type {
	case Add:
		d.applyDelete(op.Position, len(op.Data))
	case Delete:
		d.applyAdd(op.Position, op.ExistingText)
	case Replace:
		d.applyDelete(op.Position, len(op.Data))
		d.applyAdd(op.Position, op.ExistingText)
	case Format:
		d.revertFormat(op.Position, op.Length, op.ExistingMeta)
	}

	d.CurrentVersion--
}

func (d *Document) revertFormat(pos int, length int, oldMetas []NodeMetadata) {
	if d.Doc.Head == nil || len(oldMetas) == 0 || length <= 0 {
		return
	}

	var metaIdx int
	remainingLen := length
	currentPos := pos

	for remainingLen > 0 {
		node, offset := d.findNodeAtCharPos(currentPos)
		if node == nil || (offset == node.StrLen && node.Next == nil) {
			break
		}

		if offset == node.StrLen {
			node = node.Next
			if node == nil {
				break
			}
			offset = 0
		}

		if offset > 0 {
			d.splitNode(node, offset)
			node = node.Next
			offset = 0
		}

		nodeLen := node.StrLen
		if nodeLen > remainingLen {
			d.splitNode(node, remainingLen)
			nodeLen = remainingLen
		}

		// Revert to old metadata
		if metaIdx < len(oldMetas) {
			node.Metadata = make(NodeMetadata)
			for k, v := range oldMetas[metaIdx] {
				node.Metadata[k] = v
			}
			metaIdx++
		} else {
			// If we've run out of saved metadata entries, apply empty metadata
			// This handles cases where node structure changed due to splits after the format was saved
			node.Metadata = make(NodeMetadata)
		}

		remainingLen -= nodeLen
		currentPos += nodeLen
	}
}

func (d *Document) Redo() {
	if d.CurrentVersion >= len(d.History)-1 {
		return
	}
	d.CurrentVersion++
	op := d.History[d.CurrentVersion]
	switch op.Type {
	case Add:
		d.applyAdd(op.Position, op.Data)
	case Delete:
		d.applyDelete(op.Position, op.Length)
	case Replace:
		d.applyDelete(op.Position, op.Length)
		d.applyAdd(op.Position, op.Data)
	case Format:
		d.applyFormat(op.Position, op.Length, op.Metadata)
	}
}

func (d *Document) String() string {
	var sb strings.Builder
	curr := d.Doc.Head
	for curr != nil {
		sb.WriteString(curr.Content)
		curr = curr.Next
	}
	return sb.String()
}

func (d *Document) Status() {
	fmt.Printf("Document: \"%s\"\n", d.String())
	fmt.Println("Nodes:")
	curr := d.Doc.Head
	idx := 0
	for curr != nil {
		metaStr := ""
		for k, v := range curr.Metadata {
			if v {
				metaStr += k + " "
			}
		}
		fmt.Printf("  [%d] \"%s\" (len=%d, cap=%d, meta=%s)\n",
			idx, curr.Content, curr.StrLen, curr.Cap, strings.TrimSpace(metaStr))
		curr = curr.Next
		idx++
	}
}

// ── Interactive CLI ──────────────────────────────────────────────────────────

func main() {
	doc := NewDocument()
	scanner := bufio.NewScanner(os.Stdin)

	// Helper: read a line from stdin with a prompt.
	readLine := func(prompt string) string {
		fmt.Print(prompt)
		scanner.Scan()
		return scanner.Text()
	}

	// Helper: read an integer from stdin with a prompt.
	readInt := func(prompt string) int {
		val := readLine(prompt)
		n, err := strconv.Atoi(val)
		if err != nil {
			fmt.Printf("  Invalid number %q, using 0.\n", val)
			return 0
		}
		return n
	}

	fmt.Println("doc-tool — Interactive Document Editor")
	fmt.Println("Commands: insert, delete, replace, format, undo, redo, status, help, exit")

	for {
		fmt.Print("\ndoc-tool> ")
		if !scanner.Scan() {
			break
		}
		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		switch cmd {
		case "insert":
			pos := readInt("  Position: ")
			text := readLine("  Text: ")
			doc.Apply(Operation{Type: Add, Position: pos, Data: text})
			doc.Status()

		case "delete":
			pos := readInt("  Position: ")
			length := readInt("  Length: ")
			doc.Apply(Operation{Type: Delete, Position: pos, Length: length})
			doc.Status()

		case "replace":
			pos := readInt("  Position: ")
			length := readInt("  Length: ")
			text := readLine("  New text: ")
			doc.Apply(Operation{Type: Replace, Position: pos, Length: length, Data: text})
			doc.Status()

		case "format":
			pos := readInt("  Position: ")
			length := readInt("  Length: ")
			boldAns := readLine("  Bold? (y/n): ")
			italicAns := readLine("  Italic? (y/n): ")
			meta := make(NodeMetadata)
			if boldAns == "y" || boldAns == "yes" {
				meta["bold"] = true
			}
			if italicAns == "y" || italicAns == "yes" {
				meta["italic"] = true
			}
			doc.Apply(Operation{Type: Format, Position: pos, Length: length, Metadata: meta})
			doc.Status()

		case "undo":
			doc.Undo()
			doc.Status()

		case "redo":
			doc.Redo()
			doc.Status()

		case "status":
			doc.Status()

		case "help":
			fmt.Println("\n  insert  — insert text at a position")
			fmt.Println("  delete  — delete N characters at a position")
			fmt.Println("  replace — replace N characters at a position with new text")
			fmt.Println("  format  — apply bold/italic formatting to a range")
			fmt.Println("  undo    — undo the last operation")
			fmt.Println("  redo    — redo the last undone operation")
			fmt.Println("  status  — show document content and node structure")
			fmt.Println("  exit    — quit")

		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("  Unknown command %q. Type 'help' to see available commands.\n", cmd)
		}
	}
}
