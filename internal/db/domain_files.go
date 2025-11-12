package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ListFilesForDomain queries files for a domain from the database
func ListFilesForDomain(ctx context.Context, pool *pgxpool.Pool, domainID string) ([]string, error) {
	// Query the files JSONB column for the domain
	var filesJSON interface{}
	err := pool.QueryRow(ctx, `
		SELECT files
		FROM domains
		WHERE id = $1
	`, domainID).Scan(&filesJSON)

	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Extract file names from the JSONB object
	var fileNames []string

	if filesJSON != nil {
		// Convert the JSONB to a map to get the keys (filenames)
		if filesMap, ok := filesJSON.(map[string]interface{}); ok {
			for fileName := range filesMap {
				fileNames = append(fileNames, fileName)
			}
		}
	}

	// If no files found, return empty array instead of nil
	if fileNames == nil {
		fileNames = []string{}
	}

	return fileNames, nil
}
