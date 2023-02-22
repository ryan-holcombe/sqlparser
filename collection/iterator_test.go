package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIterator(t *testing.T) {
	iter := NewIterator("one", "two", "three")
	assert.Len(t, iter.slice, 3)
	assert.Equal(t, iter.length, 3)
	assert.Equal(t, -1, iter.pos)
}

func TestIterator_Next(t *testing.T) {
	iter := NewIterator("one", "two", "three")
	one, ok := iter.Next()
	assert.True(t, ok)
	assert.Equal(t, "one", one)

	two, ok := iter.Next()
	assert.True(t, ok)
	assert.Equal(t, "two", two)

	three, ok := iter.Next()
	assert.True(t, ok)
	assert.Equal(t, "three", three)

	_, ok = iter.Next()
	assert.False(t, ok)
}

func TestIterator_Peek(t *testing.T) {
	iter := NewIterator("one", "two", "three")
	one, ok := iter.Peek()
	assert.True(t, ok)
	assert.Equal(t, "one", one)
	assert.Equal(t, -1, iter.pos)
}
