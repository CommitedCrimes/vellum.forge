package cache

// lruNode represents a node in the LRU doubly-linked list
type lruNode struct {
	key  string
	prev *lruNode
	next *lruNode
}

// lruList implements a doubly-linked list for LRU tracking
type lruList struct {
	head  *lruNode
	tail  *lruNode
	nodes map[string]*lruNode
}

// newLRUList creates a new LRU list
func newLRUList() *lruList {
	head := &lruNode{}
	tail := &lruNode{}

	head.next = tail
	tail.prev = head

	return &lruList{
		head:  head,
		tail:  tail,
		nodes: make(map[string]*lruNode),
	}
}

// addToFront adds a key to the front of the list (most recently used)
func (l *lruList) addToFront(key string) {
	// Remove if already exists
	l.remove(key)

	// Create new node
	node := &lruNode{key: key}
	l.nodes[key] = node

	// Insert after head
	node.next = l.head.next
	node.prev = l.head
	l.head.next.prev = node
	l.head.next = node
}

// moveToFront moves an existing key to the front of the list
func (l *lruList) moveToFront(key string) {
	if node, exists := l.nodes[key]; exists {
		// Remove from current position
		node.prev.next = node.next
		node.next.prev = node.prev

		// Insert after head
		node.next = l.head.next
		node.prev = l.head
		l.head.next.prev = node
		l.head.next = node
	}
}

// remove removes a key from the list
func (l *lruList) remove(key string) {
	if node, exists := l.nodes[key]; exists {
		// Remove from list
		node.prev.next = node.next
		node.next.prev = node.prev

		// Remove from map
		delete(l.nodes, key)
	}
}

// removeOldest removes and returns the least recently used key
func (l *lruList) removeOldest() string {
	if l.tail.prev == l.head {
		// List is empty
		return ""
	}

	// Get the oldest node (just before tail)
	oldest := l.tail.prev
	key := oldest.key

	// Remove it
	l.remove(key)

	return key
}

// isEmpty returns true if the list is empty
func (l *lruList) isEmpty() bool {
	return l.head.next == l.tail
}
