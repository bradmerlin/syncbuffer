package syncbuffer

import (
	"sync"
	"time"
)

// Metronome is the interface that wraps the Beat and Stop methods.
type Metronome interface {
	// Beat returns a channel that emits signals.
	Beat() chan struct{}

	// Stop closes the channel returned by Beat.
	Stop()
}

// NewFrequencyMetronome returns a metronome that emits a beat at the given frequency.
// The metronome is cancelled by the given context.
func NewFrequencyMetronome(freq time.Duration) Metronome {
	m := &frequencyMetronome{
		quit: make(chan struct{}),
		freq: freq,
	}
	m.wg.Add(1)

	return m
}

type frequencyMetronome struct {
	freq time.Duration
	quit chan struct{}
	wg   sync.WaitGroup
}

func (m *frequencyMetronome) Beat() chan struct{} {
	beat := make(chan struct{})

	go func() {
		defer m.wg.Done()

		for {
			select {
			case <-m.quit:
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

func (m *frequencyMetronome) Stop() {
	close(m.quit)
	m.wg.Wait()
}

// NewManualMetronome returns a metronome and a function to manually control its beat.
func NewManualMetronome() (Metronome, func()) {

	m := manualMetronome{
		manualBeat: make(chan struct{}),
		quit:       make(chan struct{}),
	}
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()
		<-m.quit
		close(m.manualBeat)
	}()

	f := func() {
		m.manualBeat <- struct{}{}
	}

	return &m, f
}

type manualMetronome struct {
	manualBeat chan struct{}
	quit       chan struct{}
	wg         sync.WaitGroup
}

func (m *manualMetronome) Beat() chan struct{} {
	return m.manualBeat
}
func (m *manualMetronome) Stop() {
	close(m.quit)
	m.wg.Wait()
}
