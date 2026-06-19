package events

import "sync/atomic"

// Handler processes events of type T flowing through the [Chain]
// and returns a result of type R.
// A handler passes the event down through the chain by calling next,
// if it doesn't the chain execution stops there.
// If there is no next handler, next() does nothing but returns the zero
// value of type R.
type Handler[T, R any] func(event T, next func(T) R) R

// markedNode is a node in the chain that can be marked for deletion.
// This struct must be used as immutable. The only way to change the value of a markedNode is to
// create a new markedNode using [newMarkedNode].
type markedNode[T, R any] struct {
	node *handlerNode[T, R]
	// deleteMark indicates whether the node is marked for delete.
	deleteMark bool
}

func newMarkedNode[T, R any](node *handlerNode[T, R], deleted bool) *markedNode[T, R] {
	return &markedNode[T, R]{
		node:       node,
		deleteMark: deleted,
	}
}

type handlerNode[T, R any] struct {
	handler Handler[T, R]
	next    atomic.Pointer[markedNode[T, R]]
}

// Call executes the chain of nodes starting from n and returns the result.
func (n *handlerNode[T, R]) Call(event T) (result R) {
	return n.handler(event, func(e T) (r R) {
		var next = n.next.Load()
		for {
			if next == nil {
				break
			}
			if !next.deleteMark {
				break
			}
			next = next.node.next.Load()
		}
		if next == nil {
			return
		}
		return next.node.Call(e)
	})
}

// HandlerKey is a key uniquely identifying an installation of any handler in the chain,
// which can be used to remove the handler from the chain.
type HandlerKey[T, R any] struct {
	c *Chain[T, R]
	n *handlerNode[T, R]
}

// Chain is a chain of handlers processing events of type T.
type Chain[T, R any] struct {
	head atomic.Pointer[markedNode[T, R]]
}

// AddHandler adds a handler to the head of the chain.
func (c *Chain[T, R]) AddHandler(handler Handler[T, R]) HandlerKey[T, R] {
	if handler == nil {
		panic("handler cannot be nil")
	}

	for {
		head := c.head.Load()
		if head == nil {
			// the chain is empty, try add the handler as the head.
			marked := newMarkedNode(&handlerNode[T, R]{handler: handler}, false)
			if c.head.CompareAndSwap(nil, marked) {
				return HandlerKey[T, R]{c: c, n: marked.node}
			}
		} else if head.deleteMark {
			// the head is marked for delete, try update the head to the next node.
			if !c.head.CompareAndSwap(head, head.node.next.Load()) {
				continue
			}
		}
		// the chain is not empty and the head is not marked for delete, try add the handler before the head.
		marked := newMarkedNode(&handlerNode[T, R]{handler: handler}, false)
		marked.node.next.Store(head)
		if c.head.CompareAndSwap(head, marked) {
			return HandlerKey[T, R]{c: c, n: marked.node}
		}
	}
}

// RemoveHandler removes the handler identified by the key from the chain.
// True is returned if the handler identified by the key is successfully removed,
// or false if the handler is not in the chain.
//
// It is safe to call RemoveHandler during the execution of the chain.
// If the key identifies a handler of another chain, RemoveHandler panics.
func (c *Chain[T, R]) RemoveHandler(key HandlerKey[T, R]) bool {
	if key.c != c {
		panic("key is of another chain")
	}
	return c.removeNode(key.n)
}

func (c *Chain[T, R]) removeNode(node *handlerNode[T, R]) bool {
	for {
		var prev *markedNode[T, R]
		cur := c.head.Load()
		for {
			if cur == nil {
				// reach the end of the chain, the node to delete is not found, return false.
				return false
			}

			if cur.deleteMark {
				if prev == nil {
					// the head is marked for delete, try update the head to the next node.
					c.head.CompareAndSwap(cur, cur.node.next.Load())

				} else {
					// the node is not the head, try update the previous node's next to skip the marked node.
					prev.node.next.CompareAndSwap(cur, cur.node.next.Load())
				}
				// retry from the beginning.
				break
			}
			if cur.node == node {
				// found the node to delete, try mark it for delete.
				n := newMarkedNode(cur.node, true)
				if prev == nil {
					// the node to delete is the head, try mark the head for delete.
					if !c.head.CompareAndSwap(cur, n) {
						// failed to mark the head, retry from the beginning.
						break
					}

				} else {
					// the node to delete is not the head, try mark it for delete.
					if !prev.node.next.CompareAndSwap(cur, n) {
						// failed to mark the node, retry from the beginning.
						break
					}
				}
				// Do not clear the next pointer of the node to delete,
				// since it is needed for the current execution of the chain.
				return true // successfully marked the node for delete, return true.
			}
			prev = cur
			cur = cur.node.next.Load()
		}
	}
}

// Execute passes an event to the head (last added) [Handler]
// and returns the result from the handler.
// If there is no handler, the zero value of type R is returned.
// The event will flow through the chain of handlers if
// all handlers pass the event to the next handler.
func (c *Chain[T, R]) Execute(event T) (result R) {
	for {
		head := c.head.Load()
		for {
			if head == nil {
				// the chain is empty, return the zero value of type R.
				return
			}
			if head.deleteMark {
				// the head is marked for delete, try update the head to the next node.
				c.head.CompareAndSwap(head, head.node.next.Load())
				// start from the beginning
				break
			}
			// the node is not marked for delete, call the handler.
			return head.node.Call(event)
		}
	}
}
