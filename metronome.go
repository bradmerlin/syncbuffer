package syncbuffer

import (
	"context"
	"time"
)

// Metronome is the interface that wraps the Beat method.
type Metronome interface {
	// Beat returns a channel that emits signals.
	Beat() chan struct{}
}

// NewFrequencyMetronome returns a metronome that emits a beat at the given frequency.
// The metronome is cancelled by the given context.
func NewFrequencyMetronome(ctx context.Context, freq time.Duration) Metronome {
	return &frequencyMetronome{
		ctx:  ctx,
		freq: freq,
	}
}

type frequencyMetronome struct {
	ctx  context.Context
	freq time.Duration
}

func (m *frequencyMetronome) Beat() chan struct{} {
	beat := make(chan struct{})

	go func() {
		for {
			select {
			case <-m.ctx.Done():
				close(beat)
				return
			default:
				beat <- struct{}{}

				time.Sleep(m.freq)
			}
		}
	}()

	return beat
}

// NewManualMetronome returns a metronome and a function to manually control its beat.
func NewManualMetronome(ctx context.Context) (Metronome, func()) {
	manualBeat := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(manualBeat)
	}()

	m := manualMetronome{
		ctx:        ctx,
		manualBeat: manualBeat,
	}

	f := func() {
		manualBeat <- struct{}{}
	}

	return &m, f
}

type manualMetronome struct {
	ctx        context.Context
	manualBeat chan struct{}
}

func (m *manualMetronome) Beat() chan struct{} {
	return m.manualBeat
}
