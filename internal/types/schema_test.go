package types

import (
	"testing"
)

// TestStruct is a test structure for schema generation
type TestStruct struct {
	Name        string  `json:"name" jsonschema:"required,description=The name"`
	Age         int     `json:"age" jsonschema:"required,description=The age"`
	IsActive    bool    `json:"isActive" jsonschema:"description=Whether active"`
	Email       *string `json:"email,omitempty" jsonschema:"description=Email address"`
	NestedField struct {
		SubField string `json:"subField" jsonschema:"description=A sub field"`
	} `json:"nestedField" jsonschema:"description=A nested field"`
}

func TestSchemaGenerator_GenerateSchema(t *testing.T) {
	// Create schema generator
	gen := NewSchemaGenerator()

	// Create a test instance
	emailVal := "test@example.com"
	testObj := TestStruct{
		Name:     "Test",
		Age:      42,
		IsActive: true,
		Email:    &emailVal,
	}
	testObj.NestedField.SubField = "Sub value"

	// Generate schema
	schema, err := gen.GenerateSchema(testObj)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check schema
	if schema == nil {
		t.Fatalf("Generated schema is nil")
	}

	// Check type
	if typ, ok := schema["type"].(string); !ok || typ != "object" {
		t.Errorf("Expected schema type to be 'object', got %v", schema["type"])
	}

	// Check properties
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Properties not found or not a map")
	}

	// Check name property
	nameProp, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatalf("Name property not found or not a map")
	}
	if desc, ok := nameProp["description"].(string); !ok || desc != "The name" {
		t.Errorf("Expected name description to be 'The name', got %v", nameProp["description"])
	}

	// Check required fields
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("Required fields not found or not an array")
	}
	foundName := false
	foundAge := false
	for _, req := range required {
		if reqStr, ok := req.(string); ok {
			if reqStr == "name" {
				foundName = true
			}
			if reqStr == "age" {
				foundAge = true
			}
		}
	}
	if !foundName {
		t.Errorf("Required field 'name' not found")
	}
	if !foundAge {
		t.Errorf("Required field 'age' not found")
	}
}

func TestSchemaGenerator_GetArgumentNames(t *testing.T) {
	// Create schema generator
	gen := NewSchemaGenerator()

	// Create a test instance
	testObj := TestStruct{
		Name:     "Test",
		Age:      42,
		IsActive: true,
	}

	// Get argument names
	args := gen.GetArgumentNames(testObj)

	// Check arguments
	if len(args) == 0 {
		t.Fatalf("Expected at least one argument, got none")
	}

	// Check for name argument
	foundName := false
	foundAge := false
	foundIsActive := false
	for _, arg := range args {
		if arg.Name == "name" {
			foundName = true
			if !arg.Required {
				t.Errorf("Expected 'name' to be required")
			}
			if arg.Description != "The name" {
				t.Errorf("Expected 'name' description to be 'The name', got %s", arg.Description)
			}
		}
		if arg.Name == "age" {
			foundAge = true
			if !arg.Required {
				t.Errorf("Expected 'age' to be required")
			}
		}
		if arg.Name == "isActive" {
			foundIsActive = true
			if arg.Required {
				t.Errorf("Expected 'isActive' to not be required")
			}
		}
	}

	if !foundName {
		t.Errorf("Argument 'name' not found")
	}
	if !foundAge {
		t.Errorf("Argument 'age' not found")
	}
	if !foundIsActive {
		t.Errorf("Argument 'isActive' not found")
	}
}
