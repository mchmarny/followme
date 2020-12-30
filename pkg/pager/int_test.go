package pager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPager(t *testing.T) {
	list := []int64{0, 1, 2, 3, 4, 5, 6, 7}

	t.Run("nil list", func(t *testing.T) {
		_, err := GetInt64ArrayPager(nil, 10, 0)
		assert.Error(t, err)
	})

	t.Run("page size", func(t *testing.T) {
		_, err := GetInt64ArrayPager(list, 0, 0)
		assert.Error(t, err)
	})

	t.Run("page larger than list", func(t *testing.T) {
		p, err := GetInt64ArrayPager(list, 100, 0)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.True(t, p.HasNext())
		assert.False(t, p.HasPrev())
		n := p.Next()
		assert.Equal(t, len(list), len(n))
		assert.False(t, p.HasNext())
		assert.False(t, p.HasPrev())
	})

	t.Run("page equal to list", func(t *testing.T) {
		p, err := GetInt64ArrayPager(list, len(list), 0)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.True(t, p.HasNext())
		assert.False(t, p.HasPrev())
		n := p.Next()
		assert.Equal(t, len(list), len(n))
		assert.False(t, p.HasNext())
		assert.False(t, p.HasPrev())
	})

	t.Run("page loop check", func(t *testing.T) {
		p, err := GetInt64ArrayPager(list, len(list), 0)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, 0, p.GetCurrentPage())

		n := p.Next()
		assert.Equal(t, 1, p.GetCurrentPage())
		assert.Equal(t, len(list), len(n))

		n = p.Next()
		assert.Equal(t, 2, p.GetCurrentPage())
		assert.Nil(t, n)

		n = p.Next()
		assert.Equal(t, 3, p.GetCurrentPage())
		assert.Nil(t, n)
	})

	t.Run("page end to end", func(t *testing.T) {
		p, err := GetInt64ArrayPager(list, 3, 0)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.False(t, p.HasPrev())
		assert.True(t, p.HasNext())
		n := p.Next()
		assert.Equal(t, 3, len(n))
		assert.Equal(t, int64(0), n[0])
		assert.Equal(t, int64(1), n[1])
		assert.Equal(t, int64(2), n[2])
		assert.True(t, p.HasNext())
		n = p.Next()
		assert.Equal(t, 3, len(n))
		assert.Equal(t, int64(3), n[0])
		assert.Equal(t, int64(4), n[1])
		assert.Equal(t, int64(5), n[2])
		assert.True(t, p.HasNext())
		n = p.Next() // gets last 2
		assert.Equal(t, 2, len(n))
		assert.Equal(t, int64(6), n[0])
		assert.Equal(t, int64(7), n[1])
		assert.False(t, p.HasNext())
		n = p.Next() // no more left
		assert.Nil(t, n)
		p.Reset()
		n = p.Next()
		assert.Equal(t, 3, len(n))
		assert.Equal(t, int64(0), n[0])
		assert.Equal(t, int64(1), n[1])
		assert.Equal(t, int64(2), n[2])
	})

	t.Run("get N'th page", func(t *testing.T) {
		p, err := GetInt64ArrayPager(list, 3, 2)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.True(t, p.HasPrev())
		assert.True(t, p.HasNext())
		n := p.Next()
		assert.Equal(t, 2, len(n))
		assert.Equal(t, int64(6), n[0])
		assert.Equal(t, int64(7), n[1])
		assert.False(t, p.HasNext())
	})
}
