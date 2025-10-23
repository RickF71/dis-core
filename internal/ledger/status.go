package ledger

import (
	"database/sql"
	"time"
)

// DBStatus represents the current health and structure snapshot of the database.
type DBStatus struct {
	OK            bool             `json:"ok"`
	DBName        string           `json:"db_name"`
	SchemaVersion string           `json:"schema_version"`
	TableCounts   map[string]int64 `json:"table_counts"`
	Timestamp     time.Time        `json:"timestamp"`
	Message       string           `json:"message,omitempty"`
}

// GetDBStatus returns a lightweight view of the database’s current health and key stats.
func (l *Ledger) GetDBStatus() (*DBStatus, error) {
	status := &DBStatus{
		OK:            true,
		SchemaVersion: "v0.9.8",
		TableCounts:   make(map[string]int64),
		Timestamp:     time.Now(),
	}

	// Detect database name
	if l.DB != nil {
		var dbName string
		_ = l.DB.QueryRow("SELECT current_database()").Scan(&dbName)
		status.DBName = dbName
	}

	// Quick table count queries (add more as you go)
	tables := []string{"receipts", "import_receipts", "domains"}
	for _, tbl := range tables {
		var count int64
		err := l.DB.QueryRow("SELECT COUNT(*) FROM " + tbl).Scan(&count)
		if err != nil {
			if err == sql.ErrNoRows {
				count = 0
			} else {
				status.OK = false
				status.Message = "Error reading table counts"
			}
		}
		status.TableCounts[tbl] = count
	}

	return status, nil
}
