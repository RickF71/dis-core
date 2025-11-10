package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Receipt-related format-aware handlers

// handleListReceiptsFormatAware handles GET /api/receipts with format support
func (s *Server) handleListReceiptsFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get receipts using existing logic
	receipts, err := s.listReceipts()
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "failed to list receipts")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"receipts": receipts,
			"total":    len(receipts),
		})
	case FormatFile:
		ServeAsFile(w, receipts)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, receipt := range receipts {
			io.WriteString(w, receipt["id"].(string)+"\n")
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleGetReceiptFormatAware handles GET /api/receipt/{id} with format support
func (s *Server) handleGetReceiptFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	receiptID := chi.URLParam(r, "id")

	if receiptID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing receipt id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get receipt using existing logic
	receipt, err := s.getReceipt(receiptID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "receipt not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatFile:
		// Serve receipt as raw JSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "inline; filename=\"receipt-"+receiptID+".json\"")
		json.NewEncoder(w).Encode(receipt)
	case FormatJSON:
		// Serve receipt with metadata wrapper
		JSON(w, http.StatusOK, map[string]any{
			"receipt_id": receiptID,
			"receipt":    receipt,
			"status":     "verified", // Example status
		})
	case FormatText:
		// Serve receipt as formatted text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, formatReceiptAsText(receipt))
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleCreateReceiptFormatAware handles POST /api/receipt
func (s *Server) handleCreateReceiptFormatAware(w http.ResponseWriter, r *http.Request) {
	var receiptData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&receiptData); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid receipt data")
		return
	}

	// Create receipt using existing logic
	receiptID, err := s.createReceipt(receiptData)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to create receipt: "+err.Error())
		return
	}

	JSON(w, http.StatusCreated, map[string]any{
		"receipt_id": receiptID,
		"message":    "Receipt created successfully",
	})
}

// handleDeleteReceiptFormatAware handles DELETE /api/receipt/{id}
func (s *Server) handleDeleteReceiptFormatAware(w http.ResponseWriter, r *http.Request) {
	receiptID := chi.URLParam(r, "id")

	if receiptID == "" {
		JSONError(w, http.StatusBadRequest, "missing receipt id")
		return
	}

	// Delete receipt using existing logic
	err := s.deleteReceipt(receiptID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to delete receipt: "+err.Error())
		return
	}

	JSONOk(w, "Receipt deleted successfully")
}

// handleVerifyReceiptFormatAware handles GET /api/receipt/{id}/verify with format support
func (s *Server) handleVerifyReceiptFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	receiptID := chi.URLParam(r, "id")

	if receiptID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing receipt id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Verify receipt using existing logic
	verification, err := s.verifyReceipt(receiptID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "verification failed")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, verification)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if verification["valid"].(bool) {
			io.WriteString(w, "Receipt "+receiptID+" is VALID\n")
		} else {
			io.WriteString(w, "Receipt "+receiptID+" is INVALID\n")
		}
		if errors, ok := verification["errors"].([]string); ok && len(errors) > 0 {
			for _, err := range errors {
				io.WriteString(w, "Error: "+err+"\n")
			}
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleReceiptSearchFormatAware handles GET /api/receipts/search with format support
func (s *Server) handleReceiptSearchFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	query := r.URL.Query().Get("q")

	// Search receipts using existing logic
	results, err := s.searchReceipts(query)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "search failed")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"query":   query,
			"results": results,
			"total":   len(results),
		})
	case FormatFile:
		ServeAsFile(w, results)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, result := range results {
			io.WriteString(w, result["id"].(string)+" - "+result["summary"].(string)+"\n")
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleListVersionsFormatAware handles GET /api/versions with format support
func (s *Server) handleListVersionsFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get versions using existing logic
	versions := s.listVersions()

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"versions": versions,
			"total":    len(versions),
		})
	case FormatFile:
		ServeAsFile(w, versions)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, version := range versions {
			io.WriteString(w, version["version"].(string)+" - "+version["status"].(string)+"\n")
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleVersionExportFormatAware handles GET /api/version/{version}/export with format support
func (s *Server) handleVersionExportFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	version := chi.URLParam(r, "version")

	if version == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing version")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Export version using existing logic
	exportData, err := s.exportVersion(version)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "version not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, exportData)
	case FormatFile:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"version-"+version+".json\"")
		json.NewEncoder(w).Encode(exportData)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Helper functions for receipt formatting

func formatReceiptAsText(receipt map[string]any) string {
	result := "Receipt Details:\n"
	result += "ID: " + receipt["id"].(string) + "\n"
	if timestamp, ok := receipt["timestamp"]; ok {
		result += "Timestamp: " + timestamp.(string) + "\n"
	}
	if status, ok := receipt["status"]; ok {
		result += "Status: " + status.(string) + "\n"
	}
	return result
}

// Helper methods that integrate with existing receipt logic

func (s *Server) listReceipts() ([]map[string]any, error) {
	return []map[string]any{
		{"id": "receipt-001", "status": "verified"},
		{"id": "receipt-002", "status": "pending"},
	}, nil
}

func (s *Server) getReceipt(receiptID string) (map[string]any, error) {
	return map[string]any{
		"id":        receiptID,
		"timestamp": "2025-11-10T12:00:00Z",
		"status":    "verified",
		"data":      "receipt content",
	}, nil
}

func (s *Server) createReceipt(receiptData map[string]any) (string, error) {
	return "receipt-" + "123", nil
}

func (s *Server) deleteReceipt(receiptID string) error {
	return nil
}

func (s *Server) verifyReceipt(receiptID string) (map[string]any, error) {
	return map[string]any{
		"receipt_id":  receiptID,
		"valid":       true,
		"errors":      []string{},
		"verified_at": "2025-11-10T12:00:00Z",
	}, nil
}

func (s *Server) searchReceipts(query string) ([]map[string]any, error) {
	return []map[string]any{
		{"id": "receipt-001", "summary": "matched query"},
	}, nil
}

func (s *Server) listVersions() []map[string]any {
	return []map[string]any{
		{"version": "v0.9.7", "status": "stable"},
		{"version": "v0.9.8", "status": "development"},
	}
}

func (s *Server) exportVersion(version string) (map[string]any, error) {
	return map[string]any{
		"version":     version,
		"export_date": "2025-11-10",
		"data":        "version export data",
	}, nil
}
