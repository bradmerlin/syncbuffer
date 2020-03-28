package syncbuffer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncBufferOrder(t *testing.T) {
	size := 10

	sb := NewSyncBuffer(time.Millisecond, size)

	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Streamer should now start from the beginning of the buffer.
	r := NewStreamer(sb)

	var count int
	for p := range r.Stream() {
		if count < size {
			assert.Equal(t, []byte{byte(count)}, p)
		}

		count++
		if count == size {
			sb.Close()
		}
	}
}

func TestSyncBufferCursor(t *testing.T) {
	const size = 10

	sb := NewSyncBuffer(time.Millisecond, size)
	defer sb.Close()

	for i := 0; i < size+1; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Streamer should stream the entire buffer.
	s := NewStreamer(sb)

	var result int

	for p := range s.Stream() {
		r := int(p[0])
		if r == 10 {
			s.Close()
		}
		result += r
	}

	assert.Equal(t, (size*(size+1))/2, result)

	// Add 5 more
	for i := 11; i < 16; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Streamer should only stream the last 10 items.
	s = NewStreamer(sb)

	result = 0

	for p := range s.Stream() {
		r := int(p[0])
		if r == 15 {
			s.Close()
		}
		result += r
	}

	assert.Equal(t, 105, result)
}

// TestSyncBuffer_Close mainly checks for goroutine leaks.
func TestSyncBuffer_Close(t *testing.T) {
	size := 10

	sb := NewSyncBuffer(time.Millisecond, size)

	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}
	r := NewStreamer(sb)
	sb.Close()

	var result int
	for range r.Stream() {
		result++
	}
	assert.Equal(t, 0, result)
}

// TestSyncBuffer_Streamer mainly checks for goroutine leaks.
func TestSyncBuffer_Streamer(t *testing.T) {
	size := 10

	sb := NewSyncBuffer(time.Millisecond, size)
	defer sb.Close()

	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	r := NewStreamer(sb)

	r.Close()

	for range r.Stream() {
	}
}
