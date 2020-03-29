package syncbuffer

import (
	"context"
	"time"
)

// NewSyncBuffer returns a sync buffer.
func NewSyncBuffer(ctx context.Context, m Metronome) *SyncBuffer {
	r := &SyncBuffer{
		ctx: ctx,
	}

	r.startClock(m)

	return r
}

// SyncBuffer contains an internal read cursor that moves forward
// at the buffer's frequency.
type SyncBuffer struct {
	SeekBuffer
	ctx context.Context
}

// Reader returns a new Streamer that points at the parent buffer.
func (s *SyncBuffer) Reader(ctx context.Context) *Streamer {
	return &Streamer{
		t:      s,
		ctx:    ctx,
		cursor: s.Cursor(),
	}
}

// startClock moves the buffer's internal cursor forward at the metronome's frequency.
func (s *SyncBuffer) startClock(m Metronome) {
	go func() {
		for range m.Beat() {
			s.Increment()
		}
	}()
}

// Streamer is a thread-safe way to stream items from a parent buffer.
type Streamer struct {
	t      *SyncBuffer
	ctx    context.Context
	cursor int
}

// Stream returns a channel that will emit packets from the parent's current cursor
// and will wait for more packets if the end is reached.
func (sr *Streamer) Stream() chan []byte {
	output := make(chan []byte)

	go func() {
		for {
			select {
			case <-sr.ctx.Done():
				// If the caller closes, just exit immediately.
				close(output)
				return
			case <-sr.t.ctx.Done():
				// If parent closes, check if there is any data remaining.
				// If so, send the remaining data as one packet.
				remainder := sr.t.Rest(sr.cursor)
				if len(remainder) > 0 {
					output <- remainder
				}

				close(output)
				return
			default:
			}

			// If we can't read anything from the parent buffer, just wait one second.
			packet := sr.t.Read(sr.cursor)
			if packet == nil {
				time.Sleep(time.Second)
				continue
			}

			output <- packet
			sr.cursor++
		}
	}()

	return output
}
