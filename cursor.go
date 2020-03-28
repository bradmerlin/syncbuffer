package syncbuffer

import (
	"context"
	"sync"
	"time"
)

type Cursor interface {
	Position() int
}
type limitFunc func() int

func NewManualCursor(limit limitFunc) (Cursor, func()) {
	mc := &manualCursor{}
	return mc, mc.Increment
}

type manualCursor struct {
	cursor int
	limit  limitFunc
}

func (mc *manualCursor) Position() int {
	return mc.cursor
}

func (mc *manualCursor) Increment() {
	if mc.cursor == mc.limit() {
		return
	}

	mc.cursor++
}

type frequencyCursor struct {
	cursor int
	freq   time.Duration
	ctx    context.Context
	lock   sync.RWMutex
	limit  limitFunc
}

func (f *frequencyCursor) Position() int {
	f.lock.RLock()
	c := f.cursor
	f.lock.RUnlock()
	return c
}

// func (f *frequencyCursor) Stop() {
// 	f.
// }

// startClock moves the buffer's internal cursor forward at a frequency specified
// by the buffer.
func (f *frequencyCursor) startClock() {
	go func() {
		for {
			if f.cursor < f.limit() {
				select {
				case <-f.ctx.Done():
					return
				default:
				}

				f.lock.Lock()
				f.cursor++
				f.lock.Unlock()
			}
			time.Sleep(f.freq)
		}
	}()
}

func NewFrequencyCursor(ctx context.Context, freq time.Duration, limit limitFunc) Cursor {
	f := frequencyCursor{
		ctx:   ctx,
		freq:  freq,
		limit: limit,
	}
	f.startClock()

	return &f
}
