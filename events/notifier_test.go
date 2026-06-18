package events

import (
	"slices"
	"testing"
)

func TestNotifier(t *testing.T) {
	var result []int
	var notifier Notifier[int]

	notifier.Subscribe(func(event int) {
		result = append(result, event)
	})

	notifier.Subscribe(func(event int) {
		result = append(result, event+1)
	})

	notifier.Notify(1)
	notifier.Notify(10)

	expected := []int{1, 2, 10, 11}

	if !slices.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
