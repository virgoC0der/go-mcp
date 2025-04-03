package types

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
)

// SchemaGenerator generates JSON Schema from a Go struct
type SchemaGenerator struct{}

// NewSchemaGenerator creates a new SchemaGenerator instance
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{}
}

// GenerateSchema generates JSON Schema from a Go struct
func (sg *SchemaGenerator) GenerateSchema(v any) (map[string]any, error) {
	reflector := jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		ExpandedStruct:             true,
	}
	schema := reflector.Reflect(v)

	// Convert schema to map[string]any
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return nil, err
	}

	return schemaMap, nil
}

// GetArgumentNames extracts parameter names from a struct
func (sg *SchemaGenerator) GetArgumentNames(v any) []PromptArgument {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	var args []PromptArgument
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			nestedArgs := sg.GetArgumentNames(reflect.New(field.Type).Elem().Interface())
			args = append(args, nestedArgs...)
			continue
		}

		// Get json tag
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		name := strings.Split(tag, ",")[0]

		// Handle jsonschema tag
		desc := ""
		required := false
		jsTag := field.Tag.Get("jsonschema")
		if jsTag != "" {
			for _, part := range strings.Split(jsTag, ",") {
				if strings.HasPrefix(part, "description=") {
					desc = strings.TrimPrefix(part, "description=")
				}
				if part == "required" {
					required = true
				}
			}
		}

		args = append(args, PromptArgument{
			Name:        name,
			Description: desc,
			Required:    required,
		})
	}

	return args
}
