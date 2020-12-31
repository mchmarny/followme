package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuid(t *testing.T) {
	t.Run("unique", func(t *testing.T) {
		id1 := NewID()
		assert.NotNil(t, id1)
		assert.NotEmpty(t, id1)
		id2 := NewID()
		assert.NotNil(t, id2)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})
}
