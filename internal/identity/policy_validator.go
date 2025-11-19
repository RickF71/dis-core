package identity

import (
	"encoding/json"
	"fmt"
)

// ValidationError represents a single policy validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"` // "required", "type", "constraint", "unknown_field"
}

// ValidationResult contains validation outcome and any errors
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// PolicyValidator validates identity policies against schemas
type PolicyValidator struct {
	resolver *SchemaResolver
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator(resolver *SchemaResolver) *PolicyValidator {
	return &PolicyValidator{resolver: resolver}
}

// ValidatePolicy validates a policy against the effective schema for a domain
func (pv *PolicyValidator) ValidatePolicy(domainID string, policy map[string]interface{}, schemaSet *SchemaSet) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if schemaSet == nil || schemaSet.Effective == nil {
		// No schema to validate against - allow anything
		return result
	}

	effectiveProps, ok := schemaSet.Effective["properties"].(map[string]interface{})
	if !ok {
		// No properties defined - allow anything
		return result
	}

	requiredFields := pv.extractRequiredFields(schemaSet.Effective)

	// Check required fields
	for _, fieldName := range requiredFields {
		if _, exists := policy[fieldName]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Required field '%s' is missing", fieldName),
				Type:    "required",
			})
		}
	}

	// Validate each policy field
	for fieldName, fieldValue := range policy {
		fieldSchema, exists := effectiveProps[fieldName]
		if !exists {
			// Field not in schema - could warn but don't fail
			// Allows for forward compatibility
			continue
		}

		fieldSchemaMap, ok := fieldSchema.(map[string]interface{})
		if !ok {
			continue
		}

		// Validate type
		if expectedType, ok := fieldSchemaMap["type"].(string); ok {
			if err := pv.validateType(fieldName, fieldValue, expectedType); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, *err)
			}
		}

		// Validate constraints
		if errs := pv.validateConstraints(fieldName, fieldValue, fieldSchemaMap); len(errs) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, errs...)
		}
	}

	return result
}

// validateType checks if a value matches the expected JSON Schema type
func (pv *PolicyValidator) validateType(fieldName string, value interface{}, expectedType string) *ValidationError {
	actualType := pv.getJSONType(value)

	if actualType != expectedType {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' has type '%s' but schema requires '%s'", fieldName, actualType, expectedType),
			Type:    "type",
		}
	}

	return nil
}

// validateConstraints checks various JSON Schema constraints
func (pv *PolicyValidator) validateConstraints(fieldName string, value interface{}, schema map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Validate minimum (for numbers)
	if min, ok := schema["minimum"].(float64); ok {
		if numVal, ok := value.(float64); ok {
			if numVal < min {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Field '%s' value %v is less than minimum %v", fieldName, numVal, min),
					Type:    "constraint",
				})
			}
		} else if intVal, ok := value.(int); ok {
			if float64(intVal) < min {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Field '%s' value %v is less than minimum %v", fieldName, intVal, min),
					Type:    "constraint",
				})
			}
		}
	}

	// Validate maximum (for numbers)
	if max, ok := schema["maximum"].(float64); ok {
		if numVal, ok := value.(float64); ok {
			if numVal > max {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Field '%s' value %v exceeds maximum %v", fieldName, numVal, max),
					Type:    "constraint",
				})
			}
		} else if intVal, ok := value.(int); ok {
			if float64(intVal) > max {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Field '%s' value %v exceeds maximum %v", fieldName, intVal, max),
					Type:    "constraint",
				})
			}
		}
	}

	// Validate enum (for strings and other types)
	if enumValues, ok := schema["enum"].([]interface{}); ok {
		valueStr := fmt.Sprintf("%v", value)
		found := false
		for _, enumVal := range enumValues {
			if fmt.Sprintf("%v", enumVal) == valueStr {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' value '%v' is not in allowed enum values: %v", fieldName, value, enumValues),
				Type:    "constraint",
			})
		}
	}

	// Validate array items (for arrays)
	if itemsSchema, ok := schema["items"].(map[string]interface{}); ok {
		if arrayVal, ok := value.([]interface{}); ok {
			for i, item := range arrayVal {
				if itemType, ok := itemsSchema["type"].(string); ok {
					itemActualType := pv.getJSONType(item)
					if itemActualType != itemType {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("%s[%d]", fieldName, i),
							Message: fmt.Sprintf("Array item has type '%s' but schema requires '%s'", itemActualType, itemType),
							Type:    "type",
						})
					}
				}

				// Check enum for array items
				if itemEnum, ok := itemsSchema["enum"].([]interface{}); ok {
					itemStr := fmt.Sprintf("%v", item)
					found := false
					for _, enumVal := range itemEnum {
						if fmt.Sprintf("%v", enumVal) == itemStr {
							found = true
							break
						}
					}
					if !found {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("%s[%d]", fieldName, i),
							Message: fmt.Sprintf("Array item value '%v' is not in allowed enum values: %v", item, itemEnum),
							Type:    "constraint",
						})
					}
				}
			}
		}
	}

	return errors
}

// getJSONType returns the JSON Schema type string for a Go value
func (pv *PolicyValidator) getJSONType(value interface{}) string {
	switch v := value.(type) {
	case bool:
		return "boolean"
	case float64, int, int32, int64:
		return "integer"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		// Try to infer from JSON encoding
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "unknown"
		}

		var decoded interface{}
		if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
			return "unknown"
		}

		return pv.getJSONType(decoded)
	}
}

// extractRequiredFields extracts the list of required fields from a schema
func (pv *PolicyValidator) extractRequiredFields(schema map[string]interface{}) []string {
	if required, ok := schema["required"].([]interface{}); ok {
		fields := make([]string, 0, len(required))
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				fields = append(fields, fieldName)
			}
		}
		return fields
	}
	return []string{}
}

// CheckEditableFields determines which fields can be edited based on adopted schemas
// Returns map of field name -> editable (true if field is in an adopted schema)
func (pv *PolicyValidator) CheckEditableFields(schemaSet *SchemaSet) map[string]bool {
	editableFields := make(map[string]bool)

	if schemaSet == nil {
		return editableFields
	}

	// Only fields from adopted schemas are editable
	for _, schema := range schemaSet.Schemas {
		if schema.Mode != "adopted" {
			continue
		}

		properties, ok := schema.Payload["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		for fieldName := range properties {
			editableFields[fieldName] = true
		}
	}

	return editableFields
}

// GetFieldSource returns which schema (and mode) defines a field
func (pv *PolicyValidator) GetFieldSource(fieldName string, schemaSet *SchemaSet) *DomainSchema {
	if schemaSet == nil {
		return nil
	}

	// Return the first schema that defines this field (precedence order)
	for i := range schemaSet.Schemas {
		schema := &schemaSet.Schemas[i]
		properties, ok := schema.Payload["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		if _, exists := properties[fieldName]; exists {
			return schema
		}
	}

	return nil
}
