package events

// Handler processes events of type T flowing through the [Chain]
// and returns a result of type R.
// A handler passes the event down through the chain by calling next,
// if it doesn't the chain execution stops there.
// If there is no next handler, next() does nothing but returns the zero
// value of type R.
type Handler[T, R any] func(event T, next func(T) R) R

type handlerNode[T, R any] struct {
	handler Handler[T, R]
	next    *handlerNode[T, R]
	prev    *handlerNode[T, R]
}

// Call executes the chain of nodes starting from n and returns the result.
func (n *handlerNode[T, R]) Call(event T) (result R) {
	return n.handler(event, func(e T) (r R) {
		next := n.next
		if next == nil {
			return
		}
		return next.Call(e)
	})
}

// HandlerKey is a key uniquely identifying an installation of any handler in the chain,
// which can be used to remove the handler from the chain.
type HandlerKey[T, R any] struct{ n *handlerNode[T, R] }

// Chain is a chain of handlers processing events of type T.
type Chain[T, R any] struct {
	head *handlerNode[T, R]
}

// AddHandler adds a handler to the head of the chain.
func (c *Chain[T, R]) AddHandler(handler Handler[T, R]) HandlerKey[T, R] {
	if handler == nil {
		panic("handler cannot be nil")
	}
	node := &handlerNode[T, R]{handler: handler, next: c.head}
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
	return HandlerKey[T, R]{n: node}
}

// RemoveHandler removes the handler identified by the key from the chain.
// If the handler identified by the key is not in the chain, RemoveHandler does nothing.
//
// It is safe to call RemoveHandler any time, even during the execution of the chain.
func (c *Chain[T, R]) RemoveHandler(key HandlerKey[T, R]) {
	c.removeNode(key.n)
}

func (c *Chain[T, R]) removeNode(node *handlerNode[T, R]) {
	if node.prev != nil {
		// node is not the head, update the previous node's next.
		node.prev.next = node.next
	} else {
		// node is the head, update the head to the next node.
		c.head = node.next
	}
	node.prev = nil
	// do not clear node.next since it is needed for the current execution of the chain.
}

// Execute passes an event to the head (last added) [Handler]
// and returns the result from the handler.
// If there is no handler, the zero value of type R is returned.
// The event will flow through the chain of handlers if
// all handlers pass the event to the next handler.
func (c *Chain[T, R]) Execute(event T) (result R) {
	if c.head == nil {
		return
	}
	return c.head.Call(event)
}
