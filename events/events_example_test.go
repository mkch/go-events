package events_test

import (
	"fmt"

	"github.com/mkch/go-events/events"
)

func ExampleNotifier() {
	var notifier events.Notifier[string]

	notifier.Subscribe(func(event string) {
		fmt.Println("Observer 1 received:", event)
	})

	notifier.Subscribe(func(event string) {
		fmt.Println("Observer 2 received:", event)
	})

	notifier.Notify("Hello!")

	// Output:
	// Observer 1 received: Hello!
	// Observer 2 received: Hello!
}

func ExampleChain() {
	chain := events.NewChain(func(n int) bool {
		fmt.Println("Default processing of", n)
		return true
	})

	chain.AddHandler(func(event int, next func(int) bool) (ok bool) {
		// Call default processor first
		if !next(event) {
			return
		}
		// Do further processing
		fmt.Println("Further processing of", event)
		return true
	})

	chain.AddHandler(func(event int, next func(int) bool) bool {
		// Do validation first
		if event < 0 {
			fmt.Println("Invalid event", event)
			return false
		}
		// Do normal processing
		return next(event)
	})

	fmt.Println("Event 1")
	fmt.Println(chain.Execute(1))
	fmt.Println()
	fmt.Println("Event -1")
	fmt.Println(chain.Execute(-1))

	// Output:
	// Event 1
	// Default processing of 1
	// Further processing of 1
	// true
	//
	// Event -1
	// Invalid event -1
	// false
}
