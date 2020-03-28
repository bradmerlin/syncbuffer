package syncbuffer

import (
	"context"
	"sync"
	"time"
)

// NewSyncBuffer returns a cancellable streaming buffer.
func NewSyncBuffer(ctx context.Context, freq time.Duration) *SyncBuffer {
	c, canceller := context.WithCancel(ctx)

	r := &SyncBuffer{
		ctx:    c,
		cancel: canceller,
		freq:   freq,
	}

	r.startClock()

	return r
}

// SyncBuffer contains an internal read cursor that moves forward
// at the given frequency.
type SyncBuffer struct {
	data   [][]byte
	cancel context.CancelFunc
	freq   time.Duration
	mux    sync.RWMutex
	ctx    context.Context
	cursor int
}

// Add appends items to the buffer.
func (t *SyncBuffer) Add(p []byte) {
	t.mux.Lock()
	t.data = append(t.data, p)
	t.mux.Unlock()
}

// startClock moves the buffer's internal cursor forward at a frequency specified
// by the buffer.
func (t *SyncBuffer) startClock() {
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				return
			default:
			}

			if t.cursor < len(t.data) {
				t.cursor++
			}

			time.Sleep(t.freq)
		}
	}()
}

// Close stops the cursor from moving forward, and flushes all readers.
func (t *SyncBuffer) Close() {
	t.cancel()
}

// Reader returns a new StreamReader that points at the parent buffer.
func (t *SyncBuffer) Reader(ctx context.Context) *StreamReader {
	return &StreamReader{
		t:      t,
		cursor: t.cursor,
		ctx:    ctx,
	}
}

// StreamReader is a thread-safe way to stream items from a parent buffer.
type StreamReader struct {
	t      *SyncBuffer
	ctx    context.Context
	cursor int
}

// Stream returns a channel that will emit packets at the parent buffer's frequency.
func (sr *StreamReader) Stream() chan []byte {
	output := make(chan []byte)
	go func() {
		for {
			if sr.cursor >= len(sr.t.data) {
				close(output)
				return
			}

			select {
			case <-sr.t.ctx.Done():
				// If parent closes, send the remaining data as one packet.
				var remainder []byte
				for _, p := range sr.t.data[sr.cursor:] {
					remainder = append(remainder, p...)
				}

				output <- remainder

				close(output)
				return
			case <-sr.ctx.Done():
				// If the caller closes, just exit immediately.
				close(output)
				return
			default:
			}

			sr.t.mux.RLock()
			output <- sr.t.data[sr.cursor]
			sr.t.mux.RUnlock()

			sr.cursor++
		}
	}()

	return output
}
