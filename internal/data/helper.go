package data

import (
	"time"

	"github.com/mchmarny/followme/pkg/format"
)

// GetDiff returns items from b that are NOT in a
func GetDiff(a, b []int64) (diff []int64) {
	m := make(map[int64]bool)
	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

// GetDateRange returns a list of date ranges since the time
func GetDateRange(since time.Time) []time.Time {
	r := make([]time.Time, 0)
	today := format.ToISODate(time.Now().UTC())
	if format.ToISODate(since) > today {
		since = time.Now().UTC()
	}

	for {
		r = append(r, since)
		if format.ToISODate(since) >= today {
			break
		}
		since = since.AddDate(0, 0, 1)
	}
	return r
}

// Contains checks for val in list
func Contains(list []int64, val int64) bool {
	for _, item := range list {
		if item == val {
			return true
		}
	}
	return false
}
