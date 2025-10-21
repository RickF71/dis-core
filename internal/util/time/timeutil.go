package time

import (
	"time"
)

// Now returns the current UTC time.
func Now() time.Time {
	return time.Now().UTC()
}

// RFC3339Nano returns the current UTC time formatted as RFC3339Nano string.
func RFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// FormatRFC3339Nano formats a provided time.Time in RFC3339Nano UTC format.
func FormatRFC3339Nano(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// ParseRFC3339Nano parses a timestamp string into a UTC time.Time.
func ParseRFC3339Nano(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// Since returns the duration since the given time.
func Since(t time.Time) time.Duration {
	return time.Since(t)
}

// UnixMilli returns the current UTC time as milliseconds since epoch.
func UnixMilli() int64 {
	return time.Now().UTC().UnixMilli()
}
