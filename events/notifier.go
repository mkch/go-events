// Package events provides utilities for building event-driven
// systems, including a Notifier for event notification and a Chain
// for processing events through a series of handlers.
package events

// Observer observes events of type T notified by [Notifier].
type Observer[T any] func(event T)

// Notifier notifies subscribed observers about events of type T.
// The zero value of Notifier is an empty notifier ready to use,
// whose Notify method does nothing.
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
