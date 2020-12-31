package format

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		r1 := ToYesNo(true)
		assert.NotEmpty(t, r1)
		assert.Equal(t, "yes", r1)
		r2 := ToYesNo(false)
		assert.NotEmpty(t, r2)
		assert.Equal(t, "no", r2)
		assert.NotEqual(t, r1, r2)
	})

	t.Run("string", func(t *testing.T) {
		r1 := NormalizeString("A12dErCf")
		assert.NotEmpty(t, r1)
		assert.Equal(t, "a12dercf", r1)
	})

	t.Run("date", func(t *testing.T) {
		layout := "2006-01-02T15:04:05.000Z"
		val := "2020-12-30T11:45:26.371Z"
		t1, err := time.Parse(layout, val)
		assert.NoError(t, err)
		t2 := ToISODate(t1)
		assert.Equal(t, "2020-12-30", t2)
	})
}
