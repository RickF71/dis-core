package db

import (
	"time"
)

// NowRFC3339Nano returns a consistent UTC timestamp string for all DIS operations.
func NowRFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// ParseRFC3339Nano parses a timestamp string into a time.Time.
func ParseRFC3339Nano(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}
