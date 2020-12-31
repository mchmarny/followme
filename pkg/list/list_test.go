package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	list := []int64{0, 1, 2, 3, 4, 5, 6, 7}

	t.Run("contains", func(t *testing.T) {
		assert.True(t, Contains(list, 7))
		assert.False(t, Contains(list, 10))
	})

	t.Run("diff", func(t *testing.T) {
		list2 := []int64{3, 4, 5, 6, 7, 8, 9, 10}
		d := GetDiff(list, list2)
		assert.Len(t, d, 3)
		d2 := GetDiff(list2, list)
		assert.Len(t, d2, 3)
	})
}
