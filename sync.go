package syncbuffer

import (
	"sync"
	"time"
)

// NewSyncBuffer returns a sync buffer.
func NewSyncBuffer(m Metronome) *SyncBuffer {
	r := &SyncBuffer{
		quit: make(chan struct{}),
		m:    m,
	}

	r.startClock()

	return r
}

// SyncBuffer contains an internal read cursor that moves forward
// at the buffer's frequency.
type SyncBuffer struct {
	SeekBuffer
	m Metronome

	quit chan struct{}
}

// Reader returns a new Streamer that points at the parent buffer.
func (s *SyncBuffer) Reader() *Streamer {
	return &Streamer{
		sb:     s,
		cursor: s.Cursor(),
		quit:   make(chan struct{}),
	}
}

func (s *SyncBuffer) Close() {
	s.m.Stop()
	close(s.quit)
}

// startClock moves the buffer's internal cursor forward at the metronome's frequency.
func (s *SyncBuffer) startClock() {
	go func() {
		for range s.m.Beat() {
			s.Increment()
		}
	}()
}

// Streamer is a thread-safe way to stream items from a parent buffer.
type Streamer struct {
	sb     *SyncBuffer
	cursor int

	quit chan struct{}
	wg   sync.WaitGroup
}

// Stream returns a channel that will emit packets from the parent's current cursor
// and will wait for more packets if the end is reached.
func (sr *Streamer) Stream() chan []byte {
	output := make(chan []byte)

	sr.wg.Add(1)
	go func() {
		defer sr.wg.Done()

		for {
			select {
			case <-sr.quit:
				// If the caller closes, just exit immediately.
				close(output)
				return
			case <-sr.sb.quit:
				// If parent closes, check if there is any data remaining.
				// If so, send the remaining data as one packet.
				remainder := sr.sb.Rest(sr.cursor)
				if len(remainder) > 0 {
					output <- remainder
				}

				close(output)
				return
			default:
			}

			// If we can'sb read anything from the parent buffer, just wait one second.
			packet := sr.sb.Read(sr.cursor)
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

func (sr *Streamer) Close() {
	close(sr.quit)
	sr.wg.Wait()
}
