package pager

import (
	"errors"
)

// GetInt64ArrayPager configures new Int64ArrayPager
func GetInt64ArrayPager(list []int64, pageSize, page int) (*Int64ArrayPager, error) {
	if list == nil {
		list = make([]int64, 0)
	}
	if pageSize < 1 {
		return nil, errors.New("page size must be a positive number")
	}
	return &Int64ArrayPager{
		list:     list,
		pageSize: pageSize,
		page:     page,
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

// GetNextPage returns current page +1
func (p *Int64ArrayPager) GetNextPage() int {
	return p.page
}

// GetPrevPage returns current page -1
func (p *Int64ArrayPager) GetPrevPage() int {
	return p.page - 2
}

// Reset resets the cursor back to it's initial stage
func (p *Int64ArrayPager) Reset() {
	p.page = 0
}

// HasNext indicates if there is more pages
func (p *Int64ArrayPager) HasNext() bool {
	start := p.page * p.pageSize
	return start < len(p.list)
}

// HasPrev indicates if there is previous page
func (p *Int64ArrayPager) HasPrev() bool {
	return p.page > 1
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
