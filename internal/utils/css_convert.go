// css_convert.go implements CSS Interchange Bridge utilitiespackage utils

// for transparent conversion between CSS text, JSON, and DB objects.
package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"dis-core/internal/models"
)

// CSSFromText converts raw CSS bytes to DomainCSS struct
func CSSFromText(data []byte, domainID string) models.DomainCSS {
	content := string(data)
	return models.DomainCSS{
		DomainID:    domainID,
		ContentType: "text/css",
		CSSContent:  content,
		Size:        len(data),
	}
}

// CSSToText converts DomainCSS struct to raw CSS bytes
func CSSToText(css models.DomainCSS) []byte {
	return []byte(css.CSSContent)
}

// CSSFromJSON converts JSON bytes to DomainCSS struct
func CSSFromJSON(data []byte) (models.DomainCSS, error) {
	var css models.DomainCSS
	err := json.Unmarshal(data, &css)
	if err != nil {
		return css, fmt.Errorf("failed to parse CSS JSON: %w", err)
	}

	// Update size to match actual content
	css.Size = len(css.CSSContent)

	return css, nil
}

// CSSToJSON converts DomainCSS struct to JSON bytes
func CSSToJSON(css models.DomainCSS) ([]byte, error) {
	// Ensure size is accurate before marshaling
	css.Size = len(css.CSSContent)

	data, err := json.MarshalIndent(css, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CSS to JSON: %w", err)
	}

	return data, nil
}

// ValidateCSS performs basic CSS validation to prevent unsafe constructs
func ValidateCSS(content string) error {
	// Check for balanced braces
	if err := validateBalancedBraces(content); err != nil {
		return &models.CSSValidationError{
			ErrorType: "invalid_css",
			Reason:    err.Error(),
		}
	}

	// Check for potentially unsafe constructs
	if err := validateSafeCSS(content); err != nil {
		return &models.CSSValidationError{
			ErrorType: "unsafe_css",
			Reason:    err.Error(),
		}
	}

	return nil
}

// validateBalancedBraces ensures CSS braces are balanced
func validateBalancedBraces(content string) error {
	openCount := 0
	for i, char := range content {
		switch char {
		case '{':
			openCount++
		case '}':
			openCount--
			if openCount < 0 {
				return fmt.Errorf("unmatched closing brace at position %d", i)
			}
		}
	}

	if openCount > 0 {
		return fmt.Errorf("unclosed braces: %d unmatched opening braces", openCount)
	}

	return nil
}

// validateSafeCSS checks for potentially unsafe CSS constructs
func validateSafeCSS(content string) error {
	// List of potentially unsafe constructs
	unsafePatterns := []struct {
		pattern string
		reason  string
	}{
		{`@import\s+url\s*\(\s*[\"']?javascript:`, "javascript imports not allowed"},
		{`expression\s*\(`, "CSS expressions not allowed"},
		{`behavior\s*:`, "CSS behaviors not allowed"},
		{`-moz-binding`, "Mozilla bindings not allowed"},
		{`@media\s+print\s*{\s*@page\s*{\s*size\s*:\s*0`, "malformed print media"},
	}

	contentLower := strings.ToLower(content)

	for _, pattern := range unsafePatterns {
		matched, err := regexp.MatchString(pattern.pattern, contentLower)
		if err != nil {
			continue // Skip if pattern is invalid
		}
		if matched {
			return fmt.Errorf("%s", pattern.reason)
		}
	}

	return nil
}

// CalculateCSSHash calculates SHA256 hash of CSS content for round-trip verification
func CalculateCSSHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// VerifyRoundTrip verifies that CSS content is identical after text↔JSON round-trip
func VerifyRoundTrip(original models.DomainCSS) error {
	// Convert to JSON and back
	jsonData, err := CSSToJSON(original)
	if err != nil {
		return fmt.Errorf("JSON conversion failed: %w", err)
	}

	restored, err := CSSFromJSON(jsonData)
	if err != nil {
		return fmt.Errorf("JSON parsing failed: %w", err)
	}

	// Compare critical fields
	if restored.CSSContent != original.CSSContent {
		return fmt.Errorf("CSS content differs after round-trip")
	}

	if restored.DomainID != original.DomainID {
		return fmt.Errorf("domain ID differs after round-trip")
	}

	if restored.Size != len(original.CSSContent) {
		return fmt.Errorf("size differs after round-trip: got %d, expected %d",
			restored.Size, len(original.CSSContent))
	}

	// Verify content hash
	originalHash := CalculateCSSHash(original.CSSContent)
	restoredHash := CalculateCSSHash(restored.CSSContent)

	if originalHash != restoredHash {
		return fmt.Errorf("content hash differs after round-trip")
	}

	return nil
}

// ParseCSSVariables extracts CSS custom properties (--var-name: value) from content
func ParseCSSVariables(content string) map[string]string {
	variables := make(map[string]string)

	// First, remove CSS comments to avoid parsing variables in comments
	commentPattern := regexp.MustCompile(`/\*.*?\*/`)
	cleanContent := commentPattern.ReplaceAllString(content, "")

	// Regular expression to match CSS custom properties
	// Matches: --variable-name: value; (with optional spaces and semicolon)
	variablePattern := regexp.MustCompile(`--([a-zA-Z0-9-_]+)\s*:\s*([^;}\n]+)`)

	matches := variablePattern.FindAllStringSubmatch(cleanContent, -1)

	for _, match := range matches {
		if len(match) == 3 {
			varName := "--" + match[1]
			varValue := strings.TrimSpace(match[2])

			// Normalize color values
			normalizedValue := normalizeColorValue(varValue)
			variables[varName] = normalizedValue
		}
	}

	return variables
} // normalizeColorValue normalizes color values to consistent formats
func normalizeColorValue(value string) string {
	value = strings.TrimSpace(value)

	// Convert hex colors to lowercase
	if strings.HasPrefix(value, "#") {
		return strings.ToLower(value)
	}

	// Normalize rgb/rgba values by removing extra spaces
	if strings.HasPrefix(value, "rgb") {
		// Remove extra spaces around commas and parentheses
		rgbPattern := regexp.MustCompile(`\s*,\s*`)
		value = rgbPattern.ReplaceAllString(value, ", ")

		spacePattern := regexp.MustCompile(`\s*\(\s*|\s*\)\s*`)
		value = spacePattern.ReplaceAllStringFunc(value, func(s string) string {
			if strings.Contains(s, "(") {
				return "("
			}
			return ")"
		})
	}

	// Normalize hsl/hsla values similarly
	if strings.HasPrefix(value, "hsl") {
		hslPattern := regexp.MustCompile(`\s*,\s*`)
		value = hslPattern.ReplaceAllString(value, ", ")

		spacePattern := regexp.MustCompile(`\s*\(\s*|\s*\)\s*`)
		value = spacePattern.ReplaceAllStringFunc(value, func(s string) string {
			if strings.Contains(s, "(") {
				return "("
			}
			return ")"
		})
	}

	return value
}

// CSSVariableMap represents the result of CSS variable extraction
type CSSVariableMap struct {
	Count     int               `json:"count"`
	Variables map[string]string `json:"variables"`
	Hash      string            `json:"hash"`
	DomainID  string            `json:"domain_id"`
}

// ExtractCSSVariableMap extracts variables and creates a complete variable map with metadata
func ExtractCSSVariableMap(content, domainID string) CSSVariableMap {
	variables := ParseCSSVariables(content)

	// Calculate hash of the variable map for change detection
	variableHash := calculateVariableMapHash(variables)

	return CSSVariableMap{
		Count:     len(variables),
		Variables: variables,
		Hash:      variableHash,
		DomainID:  domainID,
	}
}

// calculateVariableMapHash creates a hash of the variable map for change detection
func calculateVariableMapHash(variables map[string]string) string {
	// Create a consistent string representation of the variable map
	var keys []string
	for key := range variables {
		keys = append(keys, key)
	}

	// Sort keys for consistent ordering
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Build consistent string
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(":")
		builder.WriteString(variables[key])
		builder.WriteString(";")
	}

	return CalculateCSSHash(builder.String())
}
