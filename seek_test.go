package syncbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekBuffer_Add_Read(t *testing.T) {
	sb := NewSeekBuffer()

	size := 5
	for i := 0; i < size; i++ {
		sb.Add([]byte{byte(i)})
		assert.Equal(t, []byte{byte(i)}, sb.Read(i))
	}

	assert.Nil(t, sb.Read(size))
}

func TestSeekBuffer_Cursor_Increment(t *testing.T) {
	sb := NewSeekBuffer()

	// Buffer cursor always starts at 0.
	assert.Equal(t, 0, sb.Cursor())

	// Incrementing does nothing because the buffer is empty.
	sb.Increment()
	assert.Equal(t, 0, sb.Cursor())

	// Increment still does nothing because now there is actually data at index 0.
	sb.Add([]byte{})
	sb.Increment()
	assert.Equal(t, 0, sb.Cursor())

	// Increment increments.
	sb.Add([]byte{})
	sb.Increment()
	assert.Equal(t, 1, sb.Cursor())
}

func TestSeekBuffer_Rest(t *testing.T) {
	sb := NewSeekBuffer()

	sb.Add([]byte{0})
	sb.Add([]byte{0})
	sb.Add([]byte{0})

	sb.Increment()

	assert.Len(t, sb.Rest(sb.Cursor()), 2)
}
