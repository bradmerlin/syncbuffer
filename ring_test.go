package syncbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRingBuffer_List(t *testing.T) {
	r := NewRingBuffer(4)

	result := r.list(0)
	assert.Len(t, result, 0)

	// After adding one item to the buffer, it should only ever return 1 item.
	r.Add([]byte{1})
	result = r.list(0)
	assert.Equal(t, []byte{1}, result[0].([]byte))

	result = r.list(1)
	assert.Len(t, result, 0)
}

func TestNewRingBuffer_ReadFrom(t *testing.T) {
	r := NewRingBuffer(4)

	result, cursor := r.ReadFrom(4)
	assert.Empty(t, result)
	assert.Equal(t, 0, cursor)

	r.Add([]byte{0})
	result, cursor = r.ReadFrom(4)
	assert.Empty(t, result)
	assert.Equal(t, 1, cursor)

	r.Add([]byte{1})
	r.Add([]byte{2})
	r.Add([]byte{3})
	r.Add([]byte{4})

	result, cursor = r.ReadFrom(3)
	assert.Equal(t, []byte{3}, result[0].([]byte))
	assert.Equal(t, []byte{4}, result[1].([]byte))
	assert.Equal(t, 5, cursor)

	result, cursor = r.ReadFrom(cursor)
	assert.Empty(t, result)

	r.Add([]byte{5})
	result, cursor = r.ReadFrom(cursor)
	assert.Equal(t, []byte{5}, result[0].([]byte))
	assert.Equal(t, 6, cursor)
}
func TestR_ReadFrom(t *testing.T) {
	r := NewRingBuffer(4)

	var cursor int
	var result []interface{}
	var results [][]byte

	r.Add([]byte{0})
	result, cursor = r.ReadFrom(cursor)
	for _, r := range result {
		results = append(results, r.([]byte))
	}

	r.Add([]byte{1})
	r.Add([]byte{2})
	r.Add([]byte{3})
	result, cursor = r.ReadFrom(cursor)
	for _, r := range result {
		results = append(results, r.([]byte))
	}

	r.Add([]byte{4})
	r.Add([]byte{5})
	r.Add([]byte{6})
	result, cursor = r.ReadFrom(cursor)
	for _, r := range result {
		results = append(results, r.([]byte))
	}

	assert.Equal(t, [][]byte{{0}, {1}, {2}, {3}, {4}, {5}, {6}}, results)
}
