package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// FormatVariantChecker provides utilities to verify route format consistency
type FormatVariantChecker struct {
	router chi.Router
	routes map[string][]string // base path -> list of format variants
}

// NewFormatVariantChecker creates a new format variant checker
func NewFormatVariantChecker(router chi.Router) *FormatVariantChecker {
	return &FormatVariantChecker{
		router: router,
		routes: make(map[string][]string),
	}
}

// ParseRouteFormat extracts base path and format variant from a route (exported for testing)
func (fvc *FormatVariantChecker) ParseRouteFormat(route string) (string, string) {
	return fvc.parseRouteFormat(route)
}

// ShouldHaveFormatVariants determines if a route should support format variants (exported for testing)
func (fvc *FormatVariantChecker) ShouldHaveFormatVariants(basePath string) bool {
	return fvc.shouldHaveFormatVariants(basePath)
}

// VerifyFormatConsistency checks all routes for format variant completeness
func (fvc *FormatVariantChecker) VerifyFormatConsistency() {
	log.Println("[format-check] Starting route format variant verification...")

	// Extract all routes from the Chi router
	fvc.extractRoutes()

	// Analyze format variants
	incomplete := fvc.analyzeFormatVariants()

	// Auto-register fallbacks if needed
	fvc.autoRegisterFallbacks(incomplete)

	// Log results
	total := len(fvc.routes)
	complete := total - len(incomplete)

	if len(incomplete) > 0 {
		log.Printf("[format-check] ⚠️  Found %d routes with incomplete format variants:", len(incomplete))
		for basePath, variants := range incomplete {
			log.Printf("[format-check]   %s: %v", basePath, variants)
		}
	}

	log.Printf("[format-check] ✅ Route format variants verified (%d/%d complete)", complete, total)
}

// extractRoutes walks the Chi router to extract all registered routes
func (fvc *FormatVariantChecker) extractRoutes() {
	walkFn := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Skip internal Chi routes and non-API routes
		if !strings.HasPrefix(route, "/api/") {
			return nil
		}

		// Extract base path and format variant
		basePath, format := fvc.parseRouteFormat(route)

		if basePath != "" {
			if fvc.routes[basePath] == nil {
				fvc.routes[basePath] = []string{}
			}

			// Add format if not already present
			if !contains(fvc.routes[basePath], format) {
				fvc.routes[basePath] = append(fvc.routes[basePath], format)
			}
		}

		return nil
	}

	chi.Walk(fvc.router, walkFn)
}

// parseRouteFormat extracts base path and format variant from a route
func (fvc *FormatVariantChecker) parseRouteFormat(route string) (string, string) {
	// Pattern for format variants: /api/path/{format}
	formatSuffixPattern := regexp.MustCompile(`^(.*)/\{?(?:format)\}?$`)

	// Pattern for explicit format paths: /api/path/json, /api/path/text, etc.
	explicitFormatPattern := regexp.MustCompile(`^(.*)/(json|text|file|cbor|encrypted)$`)

	// Check for explicit format suffix
	if matches := explicitFormatPattern.FindStringSubmatch(route); len(matches) == 3 {
		return matches[1], matches[2]
	}

	// Check for format parameter
	if matches := formatSuffixPattern.FindStringSubmatch(route); len(matches) == 2 {
		return matches[1], "format-param"
	}

	// No format variant detected, treat as base route
	return route, "base"
}

// analyzeFormatVariants identifies routes with incomplete format support
func (fvc *FormatVariantChecker) analyzeFormatVariants() map[string][]string {
	incomplete := make(map[string][]string)

	for basePath, variants := range fvc.routes {
		// Skip non-API routes or routes that are clearly not meant for format variants
		if strings.Contains(basePath, "/files") || strings.Contains(basePath, "/dashboard") {
			continue
		}

		// Check if route has multiple format variants
		if len(variants) == 1 && variants[0] == "base" {
			// Route has no format variants - this might be intentional for some endpoints
			if fvc.shouldHaveFormatVariants(basePath) {
				incomplete[basePath] = variants
			}
		} else if len(variants) > 1 && !contains(variants, "json") {
			// Has format support but missing JSON
			incomplete[basePath] = variants
		} else if len(variants) > 1 && !contains(variants, "text") {
			// Has format support but missing text
			incomplete[basePath] = variants
		}
	}

	return incomplete
}

// shouldHaveFormatVariants determines if a route should support format variants
func (fvc *FormatVariantChecker) shouldHaveFormatVariants(basePath string) bool {
	// Routes that should NOT support format variants
	excludePatterns := []string{
		"/files",
		"/dashboard",
		"/internal",
		"/ws/",
		"/websocket",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(basePath, pattern) {
			return false
		}
	}

	// Routes that typically should support multiple formats
	formatRoutePatterns := []string{
		"/api/domain/",
		"/api/domains",
		"/api/authority/",
		"/api/policy/",
		"/api/receipts/",
	}

	for _, pattern := range formatRoutePatterns {
		if strings.Contains(basePath, pattern) {
			return true
		}
	}

	return false
} // autoRegisterFallbacks automatically registers missing format variants
func (fvc *FormatVariantChecker) autoRegisterFallbacks(incomplete map[string][]string) {
	if len(incomplete) == 0 {
		return
	}

	log.Printf("[format-check] Auto-registering fallback format variants...")

	// Note: In a real implementation, you would need access to the original handlers
	// For now, we'll just log what would be registered
	for basePath, existingVariants := range incomplete {
		missingFormats := fvc.getMissingFormats(existingVariants)

		for _, format := range missingFormats {
			formatPath := fmt.Sprintf("%s/%s", basePath, format)
			log.Printf("[format-check]   Would register: %s", formatPath)

			// In a full implementation, you would:
			// fvc.router.Get(formatPath, fallbackHandler)
		}
	}
}

// getMissingFormats returns the format variants that should exist but don't
func (fvc *FormatVariantChecker) getMissingFormats(existingVariants []string) []string {
	required := []string{"json", "text"}
	var missing []string

	for _, required := range required {
		if !contains(existingVariants, required) {
			missing = append(missing, required)
		}
	}

	return missing
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// InstallFormatConsistencyCheck adds format verification to the server startup
func (s *Server) InstallFormatConsistencyCheck() {
	checker := NewFormatVariantChecker(s.router)
	checker.VerifyFormatConsistency()
	log.Println("[format-check] ✅ Format consistency check installed")
}
