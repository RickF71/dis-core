package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// JSON writes a structured response with a status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		log.Printf("❌ JSON encode error: %v", err)
	}
}

// Error writes a standardized error message.
func Error(w http.ResponseWriter, status int, err error) {
	JSON(w, status, map[string]any{
		"error":  err.Error(),
		"status": status,
	})
}

// JSONError writes a standardized error response with timestamp
func JSONError(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// JSONUnsupportedFormat responds with format not supported error
func JSONUnsupportedFormat(w http.ResponseWriter, format string) {
	JSONError(w, http.StatusNotAcceptable,
		fmt.Sprintf("unsupported format: %s", format))
}

// JSONSuccess writes a standardized success response
func JSONSuccess(w http.ResponseWriter, message string, data any) {
	JSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"message":   message,
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// JSONOk writes a simple success response
func JSONOk(w http.ResponseWriter, message string) {
	JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": message,
	})
}

// RespondWithFormat handles format-specific responses
func RespondWithFormat(w http.ResponseWriter, r *http.Request, data any) {
	format := DetectFormat(r)

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, data)
	case FormatFile:
		ServeAsFile(w, data)
	case FormatText:
		ServeAsText(w, data)
	case FormatCBOR:
		ServeCBOR(w, data)
	case FormatEncrypted:
		ServeEncrypted(w, data)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// ServeAsFile serves data as downloadable file
func ServeAsFile(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\"data.json\"")

	// Convert data to JSON for file download
	if err := json.NewEncoder(w).Encode(data); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to encode data for file download")
	}
}

// ServeAsText serves data as plain text
func ServeAsText(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Convert data to formatted text
	if jsonData, err := json.MarshalIndent(data, "", "  "); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to convert data to text")
	} else {
		io.WriteString(w, string(jsonData))
	}
}

// ServeCBOR serves data in CBOR format (placeholder)
func ServeCBOR(w http.ResponseWriter, data any) {
	// CBOR implementation would go here
	JSONUnsupportedFormat(w, "cbor")
}

// ServeEncrypted serves encrypted data (placeholder)
func ServeEncrypted(w http.ResponseWriter, data any) {
	// Encryption implementation would go here
	JSONUnsupportedFormat(w, "encrypted")
}
