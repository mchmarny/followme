package date

import (
	"time"

	"github.com/mchmarny/followme/pkg/format"
)

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
