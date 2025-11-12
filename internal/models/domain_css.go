package models

import (
	"fmt"
	"time"
)

// DomainCSS represents a CSS resource for a domain with interchange capabilities
// between plain CSS text, JSON, and database objects.
type DomainCSS struct {
	DomainID    string `json:"domain_id" db:"domain_id"`
	ContentType string `json:"content_type" db:"content_type"`
	CSSContent  string `json:"css_content" db:"css_content"`
	Size        int    `json:"size" db:"size"`
}

// DomainCSSHistory represents historical CSS changes for provenance tracking
type DomainCSSHistory struct {
	ID          string    `json:"id" db:"id"`
	DomainID    string    `json:"domain_id" db:"domain_id"`
	ContentType string    `json:"content_type" db:"content_type"`
	CSSContent  string    `json:"css_content" db:"css_content"`
	Size        int       `json:"size" db:"size"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
}

// CSSValidationError represents CSS validation errors with detailed context
type CSSValidationError struct {
	ErrorType string `json:"error"`
	Reason    string `json:"reason"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
}

// Error implements the error interface
func (e *CSSValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorType, e.Reason)
}
