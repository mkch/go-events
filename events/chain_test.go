package events

import (
	"slices"
	"testing"
)

func TestChain_Default(t *testing.T) {
	var chain Chain[int, string]
	// Zero value of Chain should return zero value of string.
	if r := chain.Execute(0); r != "" {
		t.Errorf("expected empty string, got %s", r)
	}
}

func TestChain(t *testing.T) {
	var chain Chain[int, string]
	chain.AddHandler(func(i int, next func(int) string) string {
		// next param of the tail handler should return zero value.
		if zero := next(i); zero != "" {
			t.Errorf("expected empty string from next, got %s", zero)
		}
		return "default"
	})
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

func TestChain_RemoveHandler(t *testing.T) {
	type void struct{}
	var chain Chain[void, void]
	key1 := chain.AddHandler(func(event void, next func(void) void) void {
		t.Error("handler1 should have been removed")
		return next(event)
	})
	handler2Called := false
	chain.AddHandler(func(event void, next func(void) void) void {
		handler2Called = true
		return next(event)
	})
	chain.AddHandler(func(event void, next func(void) void) void {
		return next(event)
	})
	key3 := chain.AddHandler(func(event void, next func(void) void) void {
		t.Error("handler3 should have been removed")
		return next(event)
	})

	chain.RemoveHandler(key1)
	chain.RemoveHandler(key3)

	if r := chain.Execute(void{}); r != (void{}) {
		t.Errorf("expected zero value of void, got %v", r)
	}
	if !handler2Called {
		t.Error("handler2 should have been called")
	}
}

func TestChain_RemoveHandler_RemoveSelf(t *testing.T) {
	type void struct{}
	var chain Chain[void, void]

	defaultHandlerCalled := false
	chain.AddHandler(func(event void, next func(void) void) void {
		defaultHandlerCalled = true
		return next(event)
	})
	handler2Called := false
	var key2 HandlerKey[void, void]
	key2 = chain.AddHandler(func(event void, next func(void) void) void {
		handler2Called = true
		// Remove itself from the chain.
		chain.RemoveHandler(key2)
		return next(event)
	})
	chain.AddHandler(func(event void, next func(void) void) void {
		return next(event)
	})

	if r := chain.Execute(void{}); r != (void{}) {
		t.Errorf("expected zero value of void, got %v", r)
	}
	if !defaultHandlerCalled {
		t.Error("default handler should have been called")
	}
	if !handler2Called {
		t.Error("handler2 should have been called")
	}

	// Execute again to verify handler2 has been removed.
	defaultHandlerCalled = false
	handler2Called = false
	if r := chain.Execute(void{}); r != (void{}) {
		t.Errorf("expected zero value of void, got %v", r)
	}
	if !defaultHandlerCalled {
		t.Error("default handler should have been called")
	}
	if handler2Called {
		t.Error("handler2 should have been removed and not be called")
	}
}

func TestChain_RemoveHandler_RemoveNext(t *testing.T) {
	var chain Chain[int, int]
	handler1Called := false
	key1 := chain.AddHandler(func(event int, next func(int) int) int {
		handler1Called = true
		return next(event)
	})
	handler2Called := false
	chain.AddHandler(func(event int, next func(int) int) int {
		handler2Called = true
		chain.RemoveHandler(key1)
		return next(event)
	})
	handler3Called := false
	chain.AddHandler(func(event int, next func(int) int) int {
		handler3Called = true
		return next(event)
	})

	chain.Execute(0)
	if !handler3Called {
		t.Error("handler3 should have been called")
	}
	if !handler2Called {
		t.Error("handler2 should have been called")
	}
	if handler1Called {
		t.Error("handler1 should have been removed and not be called")
	}
}

func TestChain_RemoveHandler_Reenter(t *testing.T) {
	type void struct{}
	var chain Chain[int, void]
	var key1 HandlerKey[int, void]
	key1 = chain.AddHandler(func(event int, next func(int) void) void {
		if event > 0 {
			// Remove itself from the chain and re-enter the chain with a new event.
			chain.RemoveHandler(key1)
			chain.Execute(event - 1)
		}
		return next(event)
	})
	calledWith := make([]int, 0)
	chain.AddHandler(func(event int, next func(int) void) void {
		calledWith = append(calledWith, event)
		return next(event)
	})

	if r := chain.Execute(2); r != (void{}) {
		t.Errorf("expected zero value of void, got %v", r)
	}
	expected := []int{2, 1}
	if !slices.Equal(calledWith, expected) {
		t.Errorf("expected calledWith to be %v, got %v", expected, calledWith)
	}
}
