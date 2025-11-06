// Package events provides utilities for building event-driven
// systems, including a Notifier for event notification and a Chain
// for processing events through a series of handlers.
package events

// Observer observes events of type T notified by [Notifier].
type Observer[T any] func(event T)

// Notifier notifies subscribed observers about events of type T.
type Notifier[T any] struct {
	observers []Observer[T]
}

// Subscribe subscribes an observer to be notified of events.
func (n *Notifier[T]) Subscribe(observer Observer[T]) {
	n.observers = append(n.observers, observer)
}

// Notify notifies all subscribed observers of an event.
// The execution order of subscribed observers is unspecified
func (n *Notifier[T]) Notify(event T) {
	for _, observer := range n.observers {
		(observer)(event)
	}
}

// Handler processes events of type T flowing through the [Chain]
// and returns a result of type R.
// A handler passes the event down through the chain by calling next,
// if it doesn't the chain execution stops there.
type Handler[T, R any] func(event T, next func(T) R) R

// Chain is a chain of handlers processing events of type T.
type Chain[T, R any] struct {
	processor func(T) R
}

// NewChain creates a new [Chain] with a default processor.
// The default processor is the last processor in the chain,
// and it is called only if all handlers pass the event down
// through the chain.
// The default processor must not be nil.
func NewChain[T, R any](defaultProcessor func(T) R) *Chain[T, R] {
	c := &Chain[T, R]{
		processor: defaultProcessor,
	}
	return c
}

// AddHandler adds a handler to the chain.
func (c *Chain[T, R]) AddHandler(handler Handler[T, R]) {
	old := c.processor
	c.processor = func(e T) R {
		return handler(e, old)
	}
}

// Execute passes an event to the last added [Handler]
// or the default processor if there is no handler.
// The event will flow through the chain of handlers down to the
// default processor if all handlers pass the event to the
// next handler.
// The returned result is from the last executed handler or
// the default processor.
func (c *Chain[T, R]) Execute(event T) (result R) {
	return c.processor(event)
}
