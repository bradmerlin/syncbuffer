package syncbuffer

import (
	"context"
	"sync"
	"time"
)

// NewSyncBuffer returns a syncbuffer.
func NewSyncBuffer(ctx context.Context, freq time.Duration) *SyncBuffer {
	c, canceller := context.WithCancel(ctx)

	r := &SyncBuffer{
		b:      newCursorBuffer(),
		ctx:    c,
		cancel: canceller,
		freq:   freq,
	}

	r.startClock()

	return r
}

// SyncBuffer contains an internal read cursor that moves forward
// at the buffer's frequency.
type SyncBuffer struct {
	b      *cursorBuffer
	cancel context.CancelFunc
	freq   time.Duration
	ctx    context.Context
}

// Add appends items to the buffer.
func (s *SyncBuffer) Add(p []byte) {
	s.b.Add(p)
}

// startClock moves the buffer's internal cursor forward at the buffer's frequency.
func (s *SyncBuffer) startClock() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
			}

			s.b.Increment()

			time.Sleep(s.freq)
		}
	}()
}

// Close stops the cursor from moving forward, and flushes all readers.
func (s *SyncBuffer) Close() {
	s.cancel()
}

// Reader returns a new StreamReader that points at the parent buffer.
func (s *SyncBuffer) Reader(ctx context.Context) *StreamReader {
	return &StreamReader{
		t:      s,
		ctx:    ctx,
		cursor: s.b.Cursor(),
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
			// If we can't read anything from the parent buffer, stop streaming.
			if sr.t.b.Read(sr.cursor) == nil {
				close(output)
				return
			}

			select {
			case <-sr.t.ctx.Done():
				// If parent closes, send the remaining data as one packet.
				output <- sr.t.b.Rest(sr.cursor)
				close(output)
				return
			case <-sr.ctx.Done():
				// If the caller closes, just exit immediately.
				close(output)
				return
			default:
			}

			output <- sr.t.b.Read(sr.cursor)
			sr.cursor++
		}
	}()

	return output
}

func newCursorBuffer() *cursorBuffer {
	return &cursorBuffer{}
}

type cursorBuffer struct {
	data     [][]byte
	dataLock sync.RWMutex

	cursor     int
	cursorLock sync.RWMutex
}

// Add appends items to the buffer.
func (cb *cursorBuffer) Add(p []byte) {
	cb.dataLock.Lock()
	cb.data = append(cb.data, p)
	cb.dataLock.Unlock()
}

// Increment increments the buffer's read cursor. If the cursor is at the end of the buffer,
// this is a no-op.
func (cb *cursorBuffer) Increment() {
	if cb.cursor >= len(cb.data)-1 {
		return
	}

	cb.cursorLock.Lock()
	cb.cursor++
	cb.cursorLock.Unlock()
}

func (cb *cursorBuffer) Cursor() int {
	cb.cursorLock.RLock()
	c := cb.cursor
	cb.cursorLock.RUnlock()
	return c
}

func (cb *cursorBuffer) Read(cursor int) []byte {
	if cursor >= len(cb.data) {
		return nil
	}
	cb.dataLock.RLock()
	p := cb.data[cursor]
	cb.dataLock.RUnlock()

	return p
}

func (cb *cursorBuffer) Rest(cursor int) []byte {
	if cursor >= len(cb.data) {
		return nil
	}
	cb.dataLock.RLock()
	pl := cb.data[cursor:]
	cb.dataLock.RUnlock()

	var packet []byte
	for _, p := range pl {
		packet = append(packet, p...)
	}

	return packet
}

func (cb *cursorBuffer) Next() []byte {
	cb.Increment()
	return cb.Read(cb.Cursor())
}
