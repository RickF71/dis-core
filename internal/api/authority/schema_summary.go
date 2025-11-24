package authorityapi

import (
	"encoding/json"
	"net/http"
	"os"
)

// SchemaSummary represents merged metadata from all schema sources
type SchemaSummary struct {
	Version    string                 `json:"version"`
	Schemas    map[string]interface{} `json:"schemas"`
	Rules      map[string]interface{} `json:"rules"`
	Thresholds map[string]interface{} `json:"thresholds"`
	Phase      string                 `json:"phase"`
}

// HandleGetSchemaSummary returns merged metadata from schema.json, ci_rules.json, thresholds.json
func HandleGetSchemaSummary(w http.ResponseWriter, r *http.Request) {
	summary := SchemaSummary{
		Version: "v1.0-dev",
		Phase:   "9A-mock",
	}

	// Load schema.json metadata
	schemaData := loadJSONFile("internal/schema/schema.json")
	if schemaData != nil {
		summary.Schemas = map[string]interface{}{
			"authority_console_schema": map[string]interface{}{
				"version":     extractVersion(schemaData),
				"title":       extractField(schemaData, "title"),
				"description": extractField(schemaData, "description"),
				"properties":  len(extractProperties(schemaData)),
			},
		}
	} else {
		summary.Schemas = map[string]interface{}{
			"authority_console_schema": map[string]interface{}{
				"version":     "0.9.3",
				"title":       "DIS Authority Console Schema",
				"description": "Mock schema data",
				"properties":  8,
			},
		}
	}

	// Load ci_rules.json metadata
	rulesData := loadJSONFile("internal/policy/ci_rules.json")
	if rulesData != nil {
		rules := extractRules(rulesData)
		summary.Rules = map[string]interface{}{
			"ci_rules": map[string]interface{}{
				"version":    extractVersion(rulesData),
				"rule_count": len(rules),
				"rules":      rules,
			},
		}
	} else {
		summary.Rules = map[string]interface{}{
			"ci_rules": map[string]interface{}{
				"version":    "1.0",
				"rule_count": 3,
				"rules":      []string{"domain_validation", "schema_compliance", "provenance_tracking"},
			},
		}
	}

	// Load thresholds.json metadata
	thresholdsData := loadJSONFile("internal/policy/thresholds.json")
	if thresholdsData != nil {
		summary.Thresholds = extractThresholds(thresholdsData)
	} else {
		summary.Thresholds = map[string]interface{}{
			"policy_evaluation": map[string]interface{}{
				"max_depth":  10,
				"timeout_ms": 300,
				"max_rules":  100,
			},
			"domain_operations": map[string]interface{}{
				"max_domains_per_user": 50,
				"max_schema_size_kb":   512,
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// Helper functions
func loadJSONFile(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return result
}

func extractVersion(data map[string]interface{}) string {
	if version, ok := data["version"].(string); ok {
		return version
	}
	return "unknown"
}

func extractField(data map[string]interface{}, field string) string {
	if value, ok := data[field].(string); ok {
		return value
	}
	return "unknown"
}

func extractProperties(data map[string]interface{}) map[string]interface{} {
	if props, ok := data["properties"].(map[string]interface{}); ok {
		return props
	}
	return map[string]interface{}{}
}

func extractRules(data map[string]interface{}) []string {
	var rules []string
	if rulesArray, ok := data["rules"].([]interface{}); ok {
		for _, rule := range rulesArray {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if name, ok := ruleMap["name"].(string); ok {
					rules = append(rules, name)
				}
			}
		}
	}
	return rules
}

func extractThresholds(data map[string]interface{}) map[string]interface{} {
	if thresholds, ok := data["thresholds"].(map[string]interface{}); ok {
		return thresholds
	}
	return map[string]interface{}{}
}
