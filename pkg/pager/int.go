package pager

import (
	"errors"
)

// GetInt64ArrayPager configures new Int64ArrayPager
func GetInt64ArrayPager(list []int64, pageSize int) (*Int64ArrayPager, error) {
	if list == nil {
		return nil, errors.New("empty list")
	}
	if pageSize < 1 {
		return nil, errors.New("page size must be a positive number")
	}
	return &Int64ArrayPager{
		list:     list,
		pageSize: pageSize,
		page:     0,
	}, nil
}

// Int64ArrayPager pages through records
type Int64ArrayPager struct {
	list     []int64
	pageSize int
	page     int
}

// GetPageSize returns the configured page size
func (p *Int64ArrayPager) GetPageSize() int {
	return p.pageSize
}

// GetCurrentPage returns current page
func (p *Int64ArrayPager) GetCurrentPage() int {
	return p.page
}

// Reset resets the cursor back to it's initial stage
func (p *Int64ArrayPager) Reset() {
	p.page = 0
}

// Next returns next page from the list
func (p *Int64ArrayPager) Next() []int64 {
	start := p.page * p.pageSize
	stop := start + p.pageSize
	p.page++

	if p.page == 1 && p.pageSize >= len(p.list) {
		// one pager
		return p.list
	}

	if start >= len(p.list) {
		// reached end
		return nil
	}

	if stop > len(p.list) {
		// stop larger than the list, trim to size
		stop = len(p.list)
	}

	return p.list[start:stop]
}
