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

func TestSeekBuffer_Resize(t *testing.T) {
	sb := NewSeekBuffer()

	// Resizing a new buffer should still yield an empty result.
	sb.Resize()
	assert.Nil(t, sb.Read(sb.Cursor()))

	sb.Add([]byte{1})
	sb.Add([]byte{2})
	sb.Add([]byte{3})

	// Resizing a buffer with items but a zero cursor should return the same buffer.
	sb.Resize()
	assert.Equal(t, []byte{1}, sb.Read(sb.Cursor()))
	assert.Len(t, sb.data, 3)

	// Incrementing the cursor then resizing the buffer should result in a buffer
	// with one item popped from the beginning.
	sb.Increment()
	sb.Resize()
	assert.Equal(t, []byte{2}, sb.Read(sb.Cursor()))
	assert.Len(t, sb.data, 2)

	// Etc.
	sb.Increment()
	assert.Equal(t, []byte{3}, sb.Read(sb.Cursor()))

	// Seeking past the end of the new buffer still yields empty slice.
	sb.Increment()
	assert.Nil(t, sb.Read(sb.Cursor()))
}
