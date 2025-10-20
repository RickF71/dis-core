package mirrorspin

import (
	"sync"
	"time"
)

var (
	startTime  = time.Now()
	lastCheck  time.Time
	eventCount int
	mu         sync.Mutex
)

// RecordActivity updates reflection metrics (call once per scan loop iteration).
func RecordActivity(count int) {
	mu.Lock()
	defer mu.Unlock()
	lastCheck = time.Now()
	eventCount += count
}

// GetStatus returns current MirrorSpin metrics as a JSON-friendly map.
func GetStatus() map[string]any {
	mu.Lock()
	defer mu.Unlock()
	return map[string]any{
		"last_check":  lastCheck.Format(time.RFC3339),
		"event_count": eventCount,
		"uptime":      time.Since(startTime).String(),
	}
}
