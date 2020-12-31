package format

import (
	"time"
)

// ToISODate formats time into 2006-01-02
func ToISODate(t time.Time) string {
	return t.Format("2006-01-02")
}
