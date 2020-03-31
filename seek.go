package syncbuffer

import (
	"sync"
)

const errSeekBeforeBuffer Error = "cannot seek past the start of the buffer"
const errSeekBeyondBuffer Error = "cannot seek past the end of the buffer"

// NewSeekBuffer returns a seekable buffer.
func NewSeekBuffer() *SeekBuffer {
	return &SeekBuffer{}
}

// SeekBuffer is a thread-safe way to store byte lists and seek through them.
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

// Resize removes redundant packets from the buffer by replacing the buffer's data
// with a new array that is the remainder of the buffer after the cursor.
func (s *SeekBuffer) Resize() {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	data := make([][]byte, len(s.data)-s.Cursor())
	copy(data, s.data[s.Cursor():])

	s.data = data
	s.seek(0)
}

func (s *SeekBuffer) seek(position int) error {
	if position < 0 {
		return errSeekBeforeBuffer
	}

	if position > len(s.data) {
		return errSeekBeyondBuffer
	}

	s.cursorLock.Lock()
	defer s.cursorLock.Unlock()

	s.cursor = position
	return nil
}
