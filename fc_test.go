package fc

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestNew tests the creation of a new FlatContext.
func TestNew(t *testing.T) {
	parent := context.Background()
	fc := New(parent)

	if fc == nil {
		t.Error("New() returned nil")
	}
}

// TestBackground tests the Background function.
func TestBackground(t *testing.T) {
	fc := Background()

	if fc == nil {
		t.Error("Background() returned nil")
	}

	// Ensure the parent context is context.Background()
	if fc.parent != context.Background() {
		t.Error("Background() parent context is not context.Background()")
	}
}

// TestWithContext tests the WithContext method.
func TestWithContext(t *testing.T) {
	fc := Background()
	newParent := context.WithValue(context.Background(), "key", "value")
	fc.WithContext(newParent)

	if fc.parent != newParent {
		t.Error("WithContext() did not update the parent context")
	}
}

// TestDeadline tests the Deadline method.
func TestDeadline(t *testing.T) {
	fc := Background()
	deadline, ok := fc.Deadline()

	if ok {
		t.Error("Background context should not have a deadline")
	}
	if !deadline.IsZero() {
		t.Error("Background context deadline should be zero")
	}

	// Test with a context that has a deadline
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	fc.WithContext(ctx)

	deadline, ok = fc.Deadline()
	if !ok {
		t.Error("WithTimeout context should have a deadline")
	}
	if deadline.IsZero() {
		t.Error("WithTimeout context deadline should not be zero")
	}
}

// TestDone tests the Done method.
func TestDone(t *testing.T) {
	fc := Background()
	done := fc.Done()

	if done == nil {
		t.Error("Done() should not return nil for a non-cancelable context")
	}

	// Test with a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	fc.WithContext(ctx)

	done = fc.Done()
	if done == nil {
		t.Error("Done() should not return nil for a cancelable context")
	}

	// Cancel the context and ensure the channel is closed
	cancel()
	select {
	case <-done:
		// Expected behavior
	case <-time.After(time.Second):
		t.Error("Done() channel was not closed after cancellation")
	}
}

// TestErr tests the Err method.
func TestErr(t *testing.T) {
	fc := Background()
	if err := fc.Err(); err != nil {
		t.Error("Background context should not have an error")
	}

	// Test with a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	fc.WithContext(ctx)
	cancel()

	if err := fc.Err(); err != context.Canceled {
		t.Error("Err() should return context.Canceled after cancellation")
	}
}

// TestValue tests the Value method.
func TestValue(t *testing.T) {
	fc := Background()

	// Test value not found
	if val := fc.Value("nonexistent"); val != nil {
		t.Error("Value() should return nil for nonexistent key")
	}

	// Test value in current context
	fc.WithValue("key1", "value1")
	if val := fc.Value("key1"); val != "value1" {
		t.Error("Value() did not return the correct value for key1")
	}

	// Test value in parent context
	parent := context.WithValue(context.Background(), "key2", "value2")
	fc.WithContext(parent)
	if val := fc.Value("key2"); val != "value2" {
		t.Error("Value() did not return the correct value for key2 from parent context")
	}
}

// TestWithValue tests the WithValue method.
func TestWithValue(t *testing.T) {
	fc := Background()

	// Add a value and verify
	fc.WithValue("key", "value")
	if val := fc.Value("key"); val != "value" {
		t.Error("WithValue() did not store the value correctly")
	}

	// Overwrite a value and verify
	fc.WithValue("key", "new_value")
	if val := fc.Value("key"); val != "new_value" {
		t.Error("WithValue() did not overwrite the value correctly")
	}
}

// TestConcurrency tests concurrent access to the FlatContext.
func TestConcurrency(t *testing.T) {
	fc := Background()
	var wg sync.WaitGroup

	// Concurrently write values
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fc.WithValue(i, i)
		}(i)
	}

	// Concurrently read values
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val := fc.Value(i)
			if val != nil && val != i {
				t.Errorf("Concurrent read/write mismatch: expected %v, got %v", i, val)
			}
		}(i)
	}

	wg.Wait()
}

// TestImmutableBehavior tests that WithValue does not modify the original context.
func TestImmutableBehavior(t *testing.T) {
	fc := Background()
	fc.WithValue("key", "value")

	// Create a new context with an additional value
	newFc := fc.WithValue("new_key", "new_value")

	// Ensure the original context is unchanged
	if val := fc.Value("new_key"); val != nil {
		t.Error("WithValue() modified the original context")
	}

	// Ensure the new context has both values
	if val := newFc.Value("key"); val != "value" {
		t.Error("New context did not retain the original value")
	}
	if val := newFc.Value("new_key"); val != "new_value" {
		t.Error("New context did not store the new value")
	}
}
