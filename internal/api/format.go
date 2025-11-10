package api

import (
	"net/http"
	"strings"
)

// Format represents the requested response format
type Format string

const (
	FormatJSON      Format = "json"
	FormatFile      Format = "file"
	FormatText      Format = "text"
	FormatCBOR      Format = "cbor"
	FormatEncrypted Format = "encrypted"
)

// DetectFormat extracts format from URL path suffix
func DetectFormat(r *http.Request) Format {
	path := r.URL.Path

	if strings.HasSuffix(path, "/json") {
		return FormatJSON
	}
	if strings.HasSuffix(path, "/file") {
		return FormatFile
	}
	if strings.HasSuffix(path, "/text") {
		return FormatText
	}
	if strings.HasSuffix(path, "/cbor") {
		return FormatCBOR
	}
	if strings.HasSuffix(path, "/encrypted") {
		return FormatEncrypted
	}

	// Default to JSON for backward compatibility
	return FormatJSON
}

// StripFormatSuffix removes format suffix from path
func StripFormatSuffix(path string) string {
	suffixes := []string{"/json", "/file", "/text", "/cbor", "/encrypted"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return strings.TrimSuffix(path, suffix)
		}
	}
	return path
}

// GetFormatSuffix returns the format suffix if present
func GetFormatSuffix(path string) string {
	suffixes := []string{"/json", "/file", "/text", "/cbor", "/encrypted"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return suffix
		}
	}
	return ""
}

// IsValidFormat checks if a format string is supported
func IsValidFormat(format string) bool {
	switch Format(format) {
	case FormatJSON, FormatFile, FormatText, FormatCBOR, FormatEncrypted:
		return true
	default:
		return false
	}
}

// GetSupportedFormats returns a list of supported format strings
func GetSupportedFormats() []string {
	return []string{
		string(FormatJSON),
		string(FormatFile),
		string(FormatText),
		string(FormatCBOR),
		string(FormatEncrypted),
	}
}
