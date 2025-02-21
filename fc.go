package fc

import (
	"context"
	"sync"
	"time"
)

// FlatContext is a flattened implementation of context.Context.
// It stores key-value pairs in a map and delegates cancellation, timeout,
// and other context-related operations to its parent context.
type FlatContext struct {
	parent context.Context             // The parent context to delegate operations to
	values map[interface{}]interface{} // A map to store key-value pairs
	mu     sync.RWMutex                // A mutex to ensure thread-safe access to the values map
}

// New creates a new flattened context.
// It takes a parent context as an argument and initializes a new FlatContext
// with an empty map for storing values.
func New(parent context.Context) *FlatContext {
	return &FlatContext{
		parent: parent,
		values: make(map[interface{}]interface{}),
	}
}

// Background returns a new FlatContext with a background parent context.
// This is useful as a root context when no specific parent context is available.
func Background() *FlatContext {
	return &FlatContext{
		parent: context.Background(),
		values: make(map[interface{}]interface{}),
	}
}

// WithContext replaces the parent context of the current FlatContext with the provided context.
// This allows dynamically changing the parent context.
func (c *FlatContext) WithContext(ctx context.Context) {
	c.parent = ctx
}

// Deadline implements the context.Context interface.
// It delegates the call to the parent context's Deadline method.
// It returns the deadline time (if set) and a boolean indicating whether a deadline is set.
func (c *FlatContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

// Done implements the context.Context interface.
// It delegates the call to the parent context's Done method.
// It returns a channel that is closed when the context is canceled or times out.
func (c *FlatContext) Done() <-chan struct{} {
	return c.parent.Done()
}

// Err implements the context.Context interface.
// It delegates the call to the parent context's Err method.
// It returns an error indicating why the context was canceled or timed out.
func (c *FlatContext) Err() error {
	return c.parent.Err()
}

// Value implements the context.Context interface.
// It first checks if the key exists in the current context's values map.
// If it does, it returns the corresponding value.
// If not, it tries to retrieve the value from the parent context.
func (c *FlatContext) Value(key interface{}) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if value, ok := c.values[key]; ok {
		return value
	}
	// If not found in the current context, look for it in the parent context.
	return c.parent.Value(key)
}

// WithValue returns a new flattened context that contains the new key-value pair.
// It locks the mutex, adds the new key-value pair to the current context's values map,
// and then returns the current context instance.
// Note: This method modifies the current context in place and returns the same instance.
// If immutability is required, consider creating a new FlatContext instance instead.
func (c *FlatContext) WithValue(key, value interface{}) *FlatContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
	return c
}
