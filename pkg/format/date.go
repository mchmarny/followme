package format

import (
	"fmt"
	"time"
)

// ToISODate formats time into 2006-01-02
func ToISODate(t time.Time) string {
	return t.Format("2006-01-02")
}

// PrettyDurationSince prints pretty duration since date
func PrettyDurationSince(a time.Time) string {
	b := time.Now().UTC()

	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)

	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}

	if month < 0 {
		month += 12
		year--
	}

	if year > 0 {
		return fmt.Sprintf("%d years, %d months, and %d days", year, month, day)
	}

	if month > 1 {
		return fmt.Sprintf("%d months and %d days", month, day)
	}

	if day == 1 {
		return fmt.Sprintf("%d day", day)
	}

	return fmt.Sprintf("%d days", day)
}
