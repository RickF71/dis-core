package utils

import (
	"testing"

	"dis-core/internal/models"
)

func TestCSSConversion(t *testing.T) {
	domainID := "test-domain"
	originalCSS := `body {
	color: red;
	background: blue;
}`

	t.Run("CSS text to struct conversion", func(t *testing.T) {
		cssData := CSSFromText([]byte(originalCSS), domainID)

		if cssData.DomainID != domainID {
			t.Errorf("Expected domain_id %s, got %s", domainID, cssData.DomainID)
		}

		if cssData.ContentType != "text/css" {
			t.Errorf("Expected content_type text/css, got %s", cssData.ContentType)
		}

		if cssData.CSSContent != originalCSS {
			t.Errorf("CSS content mismatch")
		}

		if cssData.Size != len(originalCSS) {
			t.Errorf("Expected size %d, got %d", len(originalCSS), cssData.Size)
		}
	})

	t.Run("CSS struct to text conversion", func(t *testing.T) {
		css := models.DomainCSS{
			DomainID:    domainID,
			ContentType: "text/css",
			CSSContent:  originalCSS,
			Size:        len(originalCSS),
		}

		textData := CSSToText(css)
		if string(textData) != originalCSS {
			t.Errorf("CSS text conversion failed")
		}
	})

	t.Run("JSON round-trip conversion", func(t *testing.T) {
		css := models.DomainCSS{
			DomainID:    domainID,
			ContentType: "text/css",
			CSSContent:  originalCSS,
			Size:        len(originalCSS),
		}

		// Convert to JSON
		jsonData, err := CSSToJSON(css)
		if err != nil {
			t.Fatalf("Failed to convert to JSON: %v", err)
		}

		// Convert back from JSON
		restoredCSS, err := CSSFromJSON(jsonData)
		if err != nil {
			t.Fatalf("Failed to convert from JSON: %v", err)
		}

		// Verify content matches
		if restoredCSS.DomainID != css.DomainID {
			t.Errorf("Domain ID mismatch after JSON round-trip")
		}

		if restoredCSS.CSSContent != css.CSSContent {
			t.Errorf("CSS content mismatch after JSON round-trip")
		}
	})

	t.Run("Round-trip verification", func(t *testing.T) {
		css := models.DomainCSS{
			DomainID:    domainID,
			ContentType: "text/css",
			CSSContent:  originalCSS,
			Size:        len(originalCSS),
		}

		err := VerifyRoundTrip(css)
		if err != nil {
			t.Errorf("Round-trip verification failed: %v", err)
		}
	})

	t.Run("CSS validation - valid CSS", func(t *testing.T) {
		validCSS := `body { color: red; } .header { font-size: 14px; }`

		err := ValidateCSS(validCSS)
		if err != nil {
			t.Errorf("Valid CSS failed validation: %v", err)
		}
	})

	t.Run("CSS validation - unbalanced braces", func(t *testing.T) {
		invalidCSS := `body { color: red;`

		err := ValidateCSS(invalidCSS)
		if err == nil {
			t.Error("Expected validation error for unbalanced braces")
		}

		cssErr, ok := err.(*models.CSSValidationError)
		if !ok {
			t.Errorf("Expected CSSValidationError, got %T", err)
		} else if cssErr.ErrorType != "invalid_css" {
			t.Errorf("Expected error type invalid_css, got %s", cssErr.ErrorType)
		}
	})

	t.Run("CSS hash calculation", func(t *testing.T) {
		hash1 := CalculateCSSHash(originalCSS)
		hash2 := CalculateCSSHash(originalCSS)

		if hash1 != hash2 {
			t.Error("Hash calculation should be consistent")
		}

		// Different content should produce different hash
		hash3 := CalculateCSSHash("body { color: blue; }")
		if hash1 == hash3 {
			t.Error("Different CSS should produce different hashes")
		}

		// Verify hash is not empty
		if hash1 == "" {
			t.Error("Hash should not be empty")
		}
	})
}

func TestCSSVariableExtraction(t *testing.T) {
	t.Run("Parse CSS custom properties", func(t *testing.T) {
		css := `
		:root {
			--primary-color: #007acc;
			--secondary-color: #666;
			--font-size: 16px;
			--margin: 10px 20px;
		}

		.theme {
			--bg-color: rgb(240, 240, 240);
			--text-color: hsl(0, 0%, 20%);
		}

		/* Regular properties should be ignored */
		body {
			color: red;
			background: blue;
		}
		`

		variables := ParseCSSVariables(css)

		expectedVars := map[string]string{
			"--primary-color":   "#007acc",
			"--secondary-color": "#666",
			"--font-size":       "16px",
			"--margin":          "10px 20px",
			"--bg-color":        "rgb(240, 240, 240)",
			"--text-color":      "hsl(0, 0%, 20%)",
		}

		if len(variables) != len(expectedVars) {
			t.Errorf("Expected %d variables, got %d", len(expectedVars), len(variables))
		}

		for expectedVar, expectedValue := range expectedVars {
			if actualValue, exists := variables[expectedVar]; !exists {
				t.Errorf("Expected variable %s not found", expectedVar)
			} else if actualValue != expectedValue {
				t.Errorf("Variable %s: expected %s, got %s", expectedVar, expectedValue, actualValue)
			}
		}
	})

	t.Run("Color normalization", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"#FF0000", "#ff0000"},                           // Hex to lowercase
			{"#fff", "#fff"},                                 // Already lowercase
			{"rgb(255, 0, 0)", "rgb(255, 0, 0)"},             // RGB spacing normalized
			{"rgb( 255 , 0 , 0 )", "rgb(255, 0, 0)"},         // Extra spaces removed
			{"hsl(120, 100%, 50%)", "hsl(120, 100%, 50%)"},   // HSL normalized
			{"rgba(255, 0, 0, 0.5)", "rgba(255, 0, 0, 0.5)"}, // RGBA
		}

		for _, tc := range testCases {
			result := normalizeColorValue(tc.input)
			if result != tc.expected {
				t.Errorf("Color normalization: input %s, expected %s, got %s", tc.input, tc.expected, result)
			}
		}
	})

	t.Run("Extract complete variable map", func(t *testing.T) {
		domainID := "test-domain"
		css := `
		:root {
			--primary: #000;
			--secondary: #fff;
		}
		`

		variableMap := ExtractCSSVariableMap(css, domainID)

		if variableMap.DomainID != domainID {
			t.Errorf("Expected domain_id %s, got %s", domainID, variableMap.DomainID)
		}

		if variableMap.Count != 2 {
			t.Errorf("Expected count 2, got %d", variableMap.Count)
		}

		if len(variableMap.Variables) != 2 {
			t.Errorf("Expected 2 variables, got %d", len(variableMap.Variables))
		}

		if variableMap.Hash == "" {
			t.Error("Variable map hash should not be empty")
		}

		// Verify specific variables
		if variableMap.Variables["--primary"] != "#000" {
			t.Errorf("Expected --primary: #000, got %s", variableMap.Variables["--primary"])
		}

		if variableMap.Variables["--secondary"] != "#fff" {
			t.Errorf("Expected --secondary: #fff, got %s", variableMap.Variables["--secondary"])
		}
	})

	t.Run("Variable map hash consistency", func(t *testing.T) {
		css1 := `:root { --color: red; --size: 10px; }`
		css2 := `:root { --size: 10px; --color: red; }` // Different order

		map1 := ExtractCSSVariableMap(css1, "test")
		map2 := ExtractCSSVariableMap(css2, "test")

		// Should have same hash despite different order in CSS
		if map1.Hash != map2.Hash {
			t.Error("Variable map hashes should be consistent regardless of CSS order")
		}

		// Different content should produce different hash
		css3 := `:root { --color: blue; --size: 10px; }`
		map3 := ExtractCSSVariableMap(css3, "test")

		if map1.Hash == map3.Hash {
			t.Error("Different variable values should produce different hashes")
		}
	})

	t.Run("Empty CSS handling", func(t *testing.T) {
		variableMap := ExtractCSSVariableMap("", "test")

		if variableMap.Count != 0 {
			t.Errorf("Expected count 0 for empty CSS, got %d", variableMap.Count)
		}

		if len(variableMap.Variables) != 0 {
			t.Errorf("Expected empty variables map, got %d variables", len(variableMap.Variables))
		}
	})

	t.Run("Invalid variable syntax handling", func(t *testing.T) {
		css := `
		:root {
			color: red; /* Not a custom property */
			--valid: blue;
			-invalid: green; /* Invalid - single dash */
			/* --commented: yellow; */ /* Commented out */
		}
		`

		variables := ParseCSSVariables(css)

		if len(variables) != 1 {
			t.Errorf("Expected 1 valid variable, got %d", len(variables))
		}

		if variables["--valid"] != "blue" {
			t.Errorf("Expected --valid: blue, got %s", variables["--valid"])
		}

		// Should not contain invalid variables
		if _, exists := variables["color"]; exists {
			t.Error("Regular CSS property should not be treated as variable")
		}

		if _, exists := variables["-invalid"]; exists {
			t.Error("Invalid variable syntax should be ignored")
		}
	})
}
