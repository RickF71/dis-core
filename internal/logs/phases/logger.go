package phases

import (
	"fmt"
	"os"
	"time"
)

func WritePhaseLog(phase, summary string) error {
	path := fmt.Sprintf("internal/logs/phases/%s.log", phase)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	_, err = fmt.Fprintf(f, "# %s\n**Timestamp:** %s\n\n%s\n\n---\n", phase, timestamp, summary)
	return err
}
