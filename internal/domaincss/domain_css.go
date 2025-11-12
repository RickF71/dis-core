// domain_css.go implements database operations for DomainCSS with history trackingpackage domaincss

package domaincss

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/models"
)

// Store stores or updates DomainCSS with upsert functionality and maintains history
func Store(ctx context.Context, db *pgxpool.Pool, domainID string, css models.DomainCSS, updatedBy string) error {
	// Start transaction for atomic operations
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Store history record first
	historyID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_css_history (
			id, domain_id, content_type, css_content, size, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, historyID, css.DomainID, css.ContentType, css.CSSContent, css.Size, time.Now(), updatedBy)
	if err != nil {
		return fmt.Errorf("failed to store CSS history: %w", err)
	}

	// Upsert current CSS record
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_css (domain_id, content_type, css_content, size, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (domain_id)
		DO UPDATE SET
			content_type = EXCLUDED.content_type,
			css_content = EXCLUDED.css_content,
			size = EXCLUDED.size,
			updated_at = EXCLUDED.updated_at
	`, css.DomainID, css.ContentType, css.CSSContent, css.Size, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store CSS: %w", err)
	}

	// Also update the normalized domains table CSS structure (Phase 10J synchronization)
	_, err = tx.Exec(ctx, `
		UPDATE domains
		SET data = jsonb_set(
			jsonb_set(
				COALESCE(data, '{}'::jsonb),
				'{data,css}',
				jsonb_build_object(
					'content', $2::text,
					'hash', '',
					'verified', true
				),
				true
			),
			'{data,css,updated_at}',
			to_jsonb(now()::text),
			true
		)
		WHERE id = $1::uuid
	`, css.DomainID, css.CSSContent)
	if err != nil {
		return fmt.Errorf("failed to update domains table CSS: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Get retrieves DomainCSS by domain ID
func Get(ctx context.Context, db *pgxpool.Pool, domainID string) (models.DomainCSS, error) {
	var css models.DomainCSS

	err := db.QueryRow(ctx, `
		SELECT domain_id, content_type, css_content, size
		FROM domain_css WHERE domain_id = $1
	`, domainID).Scan(&css.DomainID, &css.ContentType, &css.CSSContent, &css.Size)

	if err != nil {
		return css, fmt.Errorf("failed to get CSS for domain %s: %w", domainID, err)
	}

	return css, nil
}

// GetHistory retrieves CSS history for a domain
func GetHistory(ctx context.Context, db *pgxpool.Pool, domainID string, limit int) ([]models.DomainCSSHistory, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(ctx, `
		SELECT id, domain_id, content_type, css_content, size, updated_at, updated_by
		FROM domain_css_history
		WHERE domain_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`, domainID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get CSS history: %w", err)
	}
	defer rows.Close()

	var history []models.DomainCSSHistory
	for rows.Next() {
		var h models.DomainCSSHistory
		err := rows.Scan(&h.ID, &h.DomainID, &h.ContentType, &h.CSSContent, &h.Size, &h.UpdatedAt, &h.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CSS history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// Delete removes CSS for a domain (keeps history)
func Delete(ctx context.Context, db *pgxpool.Pool, domainID string) error {
	_, err := db.Exec(ctx, `DELETE FROM domain_css WHERE domain_id = $1`, domainID)
	if err != nil {
		return fmt.Errorf("failed to delete CSS for domain %s: %w", domainID, err)
	}
	return nil
}

// CreateTables creates the necessary database tables if they don't exist
func CreateTables(ctx context.Context, db *pgxpool.Pool) error {
	// Create main domain_css table
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_css (
			domain_id    TEXT PRIMARY KEY,
			content_type TEXT NOT NULL DEFAULT 'text/css',
			css_content  TEXT NOT NULL DEFAULT '',
			size         INTEGER NOT NULL DEFAULT 0,
			updated_at   TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create domain_css table: %w", err)
	}

	// Create history table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_css_history (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id    TEXT NOT NULL,
			content_type TEXT NOT NULL,
			css_content  TEXT NOT NULL,
			size         INTEGER NOT NULL,
			updated_at   TIMESTAMPTZ NOT NULL,
			updated_by   TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create domain_css_history table: %w", err)
	}

	// Create indexes
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_domain_css_history_domain_id ON domain_css_history(domain_id);
		CREATE INDEX IF NOT EXISTS idx_domain_css_history_updated_at ON domain_css_history(updated_at DESC);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}
