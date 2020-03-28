// Package syncbuffer provides functionality for writing byte streams and keeping
// readers of those streams roughly in sync with each other.
package syncbuffer

import (
	"sync"
	"time"
)

// NewSyncBuffer returns a buffer of the given size that writes items at a
// period defined by the given frequency.
// Close must be called on the buffer after it's done with.
func NewSyncBuffer(freq time.Duration, size int) *SyncBuffer {
	sb := SyncBuffer{
		freq: freq,
		r:    NewRingBuffer(size),
		quit: make(chan struct{}),
	}
	return &sb
}

// SyncBuffer is a fixed-size buffer that allows writes at a period specified
// by its frequency. Writes via Add, and reads via a Streamer are thread-safe.
type SyncBuffer struct {
	r    *RingBuffer
	freq time.Duration

	packets [][]byte
	lock    sync.RWMutex

	quit chan struct{}
	wg   sync.WaitGroup
}

// Add then adds the given item to the buffer, overwriting the oldest item
// if the buffer is full.
// Will block for at least the period defined by the buffer's frequency.
func (sb *SyncBuffer) Add(item []byte) {
	sb.r.Add(item)
	time.Sleep(sb.freq)
}

// Close cleans up the buffer's internal goroutines and closes all streamers.
func (sb *SyncBuffer) Close() {
	close(sb.quit)
	sb.wg.Wait()
}

// NewStreamer returns a thing that allows data to be read from the parent buffer.
// Streamer.Close must be called when the streamer is done with; however, closing
// the parent buffer will have the same effect from the streamer's perspective.
func NewStreamer(sb *SyncBuffer) *Streamer {
	return &Streamer{
		sb:     sb,
		output: make(chan []byte, len(sb.r.items)),
		quit:   make(chan struct{}),
	}
}

// Streamer allows items to be read from its parent buffer.
type Streamer struct {
	sb *SyncBuffer

	output chan []byte

	quit chan struct{}
	wg   sync.WaitGroup
}

// Stream returns a channel that emits items from the parent buffer as soon
// as they become available.
func (st *Streamer) Stream() chan []byte {

	st.wg.Add(1)
	st.sb.wg.Add(1)
	go func() {
		defer st.wg.Done()
		defer st.sb.wg.Done()

		var packets []interface{}

		cursor := st.sb.r.OldestCursor()

		for {
			select {
			case <-st.sb.quit:
				close(st.output)
				return
			case <-st.quit:
				close(st.output)
				return
			default:
			}

			packets, cursor = st.sb.r.ReadFrom(cursor)

			if packets == nil {
				// If there are no packets available it means we're at the end of the list.
				// Wait for a bit.
				time.Sleep(st.sb.freq / 10)
				continue
			}
			for _, p := range packets {
				st.output <- p.([]byte)
			}
		}
	}()

	return st.output
}

// Close stops the streamer.
func (st *Streamer) Close() {
	close(st.quit)
	st.wg.Wait()
}
