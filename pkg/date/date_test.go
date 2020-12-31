package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	t.Run("since", func(t *testing.T) {
		list := GetDateRange(time.Now().UTC().AddDate(0, 0, -2))
		assert.NotNil(t, list)
		assert.Len(t, list, 3)
	})
}
