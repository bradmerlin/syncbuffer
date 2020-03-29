package syncbuffer

import (
	"sync"
)

const errSeekBeforeBuffer Error = "cannot seek past the start of the buffer"
const errSeekBeyondBuffer Error = "cannot seek past the end of the buffer"

// NewSeekBuffer returns a buffer that allows streaming from any point in its saved data.
// SeekBuffers are thread-safe.
// This buffer currently holds all information in memory.
func NewSeekBuffer() *SeekBuffer {
	return &SeekBuffer{}
}

type SeekBuffer struct {
	data     [][]byte
	dataLock sync.RWMutex

	cursor     int
	cursorLock sync.RWMutex
}

// Add appends items to the buffer.
func (s *SeekBuffer) Add(p ...[]byte) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	s.data = append(s.data, p...)
}

// Increment increments the buffer's read cursor.
// If the cursor is at the end of the buffer, this is a no-op.
func (s *SeekBuffer) Increment() {
	err := s.seek(s.Cursor() + 1)
	if err != nil {
		// No-op for now.
	}
}

func (s *SeekBuffer) Cursor() int {
	s.cursorLock.RLock()
	defer s.cursorLock.RUnlock()

	return s.cursor
}

func (s *SeekBuffer) Read(cursor int) []byte {
	if cursor >= len(s.data) {
		return nil
	}

	s.dataLock.RLock()
	defer s.dataLock.RUnlock()

	return s.data[cursor]
}

func (s *SeekBuffer) Rest(cursor int) []byte {
	if cursor >= len(s.data) {
		return nil
	}

	s.dataLock.RLock()
	pl := s.data[cursor:]
	s.dataLock.RUnlock()

	var packet []byte
	for _, p := range pl {
		packet = append(packet, p...)
	}

	return packet
}

func (s *SeekBuffer) seek(position int) error {
	if position < 0 {
		return errSeekBeforeBuffer
	}

	if position > len(s.data)-1 {
		return errSeekBeyondBuffer
	}

	s.cursorLock.Lock()
	defer s.cursorLock.Unlock()

	s.cursor = position
	return nil
}
