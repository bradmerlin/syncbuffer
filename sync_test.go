package syncbuffer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncBufferOrder(t *testing.T) {
	m, _ := NewManualMetronome()

	sb := NewSyncBuffer(m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Reader should now start from the beginning of the buffer.
	r := sb.Reader()

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
	m, beat := NewManualMetronome()
	sb := NewSyncBuffer(m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Move the parent cursor forward a bit.
	for i := 0; i < size/2; i++ {
		beat()
	}

	// I have made peace with this for now.
	time.Sleep(time.Millisecond)

	// Reader should now start from the middle of the buffer
	r := sb.Reader()

	var count int
	for range r.Stream() {
		count++
		if count == size/2 {
			sb.Close()
		}
	}

	assert.Equal(t, size/2, count)
}

func TestSyncBufferCancel(t *testing.T) {
	m, _ := NewManualMetronome()

	sb := NewSyncBuffer(m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	r := sb.Reader()

	var count int
	sb.Close()
	for p := range r.Stream() {
		if count == 2 {
			assert.Len(t, p, 8)
		}

		count++
	}
}

func TestSyncBuffer_ReaderCancel(t *testing.T) {
	m, _ := NewManualMetronome()

	sb := NewSyncBuffer(m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	r := sb.Reader()

	var count int
	for p := range r.Stream() {
		if count == 1 {
			r.Close()
		}

		if count == 2 {
			// When reader context is cancelled, it just returns its current item.
			assert.Len(t, p, 1)
		}

		count++
	}
	// When reader context is cancelled, loop just ends.
	assert.Equal(t, 2, count)
}
