package syncbuffer

import (
	"sync"
)

// NewRingBuffer returns a ring buffer of the given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		items: make([]interface{}, size),
	}
}

// RingBuffer stores a fixed-size list of any type of items.
// If the buffer is full, the oldest item is overwritten.
type RingBuffer struct {
	count int
	items []interface{}

	lock sync.RWMutex
}

// ReadFrom returns all items newer than the given cursor, and a new cursor,
// which successive reads should use to query the buffer to keep data contiguous.
// If the given cursor is out of bounds, ReadFrom will return a nil slice.
// If the item at cursor is no longer available, ReadFrom will return the
// current entire buffer.
func (r *RingBuffer) ReadFrom(cursor int) ([]interface{}, int) {
	if cursor < 0 {
		return nil, 0
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	// This is how we will signal to callers to back off.
	if cursor >= r.count {
		return nil, r.count
	}

	// If cursor is valid, but is older than the oldest item in the buffer,
	// just return the entire buffer.
	oldestCursor := r.OldestCursor()
	if cursor < oldestCursor {
		return r.list(oldestCursor), r.count
	}

	return r.list(cursor), r.count
}

// Add appends an item to the buffer.
// It may overwrite the oldest item in the buffer.
func (r *RingBuffer) Add(item interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.items[r.index(r.count)] = item

	r.count++
}

// OldestCursor returns the cursor of the oldest item in the buffer.
func (r *RingBuffer) OldestCursor() int {
	r.lock.RLock()
	defer r.lock.RUnlock()

	oldestCursor := r.count - len(r.items)
	if oldestCursor < 0 {
		oldestCursor = 0
	}

	return oldestCursor
}

// List returns all available items ahead of the cursor, from oldest to newest.
func (r *RingBuffer) list(cursor int) []interface{} {
	size := r.count - cursor

	// If the cursor is too far ahead, exit.
	if size <= 0 {
		return nil
	}

	items := make([]interface{}, 0, size)

	r.lock.RLock()
	defer r.lock.RUnlock()

	for cursor < r.count {
		items = append(items, r.items[r.index(cursor)])
		cursor++
	}

	return items
}

// Index returns the buffer index represented by a cursor.
func (r *RingBuffer) index(cursor int) int {
	return cursor % len(r.items)
}
