package syncbuffer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncBufferOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m, _ := NewManualMetronome(ctx)

	sb := NewSyncBuffer(ctx, m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Reader should now start from the beginning of the buffer.
	r := sb.Reader(ctx)

	var count int
	for p := range r.Stream() {
		assert.Equal(t, []byte{byte(count)}, p)

		count++

		if count == size {
			cancel()
		}
	}
}

func TestSyncBufferCursor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m, beat := NewManualMetronome(ctx)

	sb := NewSyncBuffer(ctx, m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	// Move the parent cursor forward a bit.
	for i := 0; i < size/2; i++ {
		beat()
	}

	// Reader should now start from the middle of the buffer
	r := sb.Reader(ctx)

	var count int
	for range r.Stream() {
		count++
		if count == size/2-1 {
			cancel()
		}
	}

	// Cancel might take a few iterations to propagate, so don't test exact values.
	assert.GreaterOrEqual(t, count, size/2-1)
	assert.Less(t, count, size/2+2)
}

func TestSyncBufferCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m, _ := NewManualMetronome(ctx)

	sb := NewSyncBuffer(ctx, m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	readerCtx, _ := context.WithCancel(ctx)
	r := sb.Reader(readerCtx)

	var count int
	for p := range r.Stream() {
		cancel()
		if count == 2 {
			assert.Len(t, p, 8)
		}

		count++
	}
}

func TestSyncBuffer_ReaderCancel(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	m, _ := NewManualMetronome(ctx)

	sb := NewSyncBuffer(ctx, m)

	size := 10
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
	}

	readerCtx, cancel := context.WithCancel(ctx)
	r := sb.Reader(readerCtx)

	var count int
	for p := range r.Stream() {
		// This is lame, but if cancel takes two iterations to propagate, it's OK.
		time.Sleep(time.Millisecond)
		cancel()

		if count == 2 {
			// When reader context is cancelled, it just returns its current item.
			assert.Len(t, p, 1)
		}

		count++
	}
	// When reader context is cancelled, loop just ends.
	assert.Equal(t, 2, count)
}
