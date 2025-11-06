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

func TestChain_Default(t *testing.T) {
	var chain = NewChain(func(i int) string { return "default" })
	if r := chain.Execute(0); r != "default" {
		t.Errorf("expected default, got %s", r)
	}
}

func TestChain(t *testing.T) {
	var chain = NewChain(func(i int) string { return "default" })
	var result []int

	chain.AddHandler(func(event int, next func(int) string) string {
		result = append(result, event)
		return "handler1-" + next(event)
	})
	chain.AddHandler(func(event int, next func(int) string) (r string) {
		r = "handler2-" + next(event)
		result = append(result, event*10)
		result = append(result, event*100)
		return
	})

	if r := chain.Execute(1); r != "handler2-handler1-default" {
		t.Errorf("expected handler2-handler1-default, got %s", r)
	}
	if !slices.Equal(result, []int{1, 10, 100}) {
		t.Errorf("expected [10 1], got %v", result)
	}
}
