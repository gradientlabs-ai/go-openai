// Package jsonschema provides very simple functionality for representing a JSON schema as a
// (nested) struct. This struct can be used with the chat completion "function call" feature.
// For more complicated schemas, it is recommended to use a dedicated JSON schema library
// and/or pass in the schema in []byte format.
package jsonschema

import (
	"encoding/json"
	"reflect"
	"strings"
)

type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// Definition is a struct for describing a JSON Schema.
// It is fairly limited, and you may have better luck using a third-party library.
type Definition struct {
	// Type specifies the data type of the schema.
	Type DataType `json:"type,omitempty"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// one element, where each element is unique. You will probably only use this with strings.
	Enum []string `json:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is Object.
	Properties map[string]Definition `json:"properties"`
	// Required specifies which properties are required, if the schema type is Object.
	Required []string `json:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is Array.
	Items *Definition `json:"items,omitempty"`
	// AdditionalProperties specifies whether additional properties are allowed.
	// When using structured outputs with strict mode, this must be set to false.
	// Use a pointer so that false is explicitly serialized (vs omitted when not set).
	AdditionalProperties *bool `json:"additionalProperties,omitempty"`
}

func (d Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(d),
	})
}

// GenerateSchemaForType generates a JSON schema Definition from a Go type using reflection.
// It supports struct tags for customization:
//   - `json:"field_name"` for the JSON field name (use "-" to skip)
//   - `jsonschema_description:"..."` for field descriptions
//   - `jsonschema_enum:"val1,val2,val3"` for enum values
//
// For OpenAI structured outputs with strict mode, set Strict: true in the response format
// and ensure AdditionalProperties is set to false at each object level.
//
// Example:
//
//	type MathResponse struct {
//	    Steps       []Step `json:"steps" jsonschema_description:"List of solution steps"`
//	    FinalAnswer string `json:"final_answer" jsonschema_description:"The final answer"`
//	}
//	schema, err := jsonschema.GenerateSchemaForType(MathResponse{})
func GenerateSchemaForType(v any) (Definition, error) {
	return generateSchema(reflect.TypeOf(v), make(map[reflect.Type]bool))
}

func generateSchema(t reflect.Type, seen map[reflect.Type]bool) (Definition, error) {
	// Handle pointers by dereferencing
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for cycles to prevent infinite recursion
	if t.Kind() == reflect.Struct && seen[t] {
		// Return empty object for recursive types
		return Definition{Type: Object, Properties: map[string]Definition{}}, nil
	}

	switch t.Kind() {
	case reflect.Struct:
		return generateStructSchema(t, seen)
	case reflect.Slice, reflect.Array:
		return generateArraySchema(t, seen)
	case reflect.Map:
		return generateMapSchema(t, seen)
	case reflect.String:
		return Definition{Type: String}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Definition{Type: Integer}, nil
	case reflect.Float32, reflect.Float64:
		return Definition{Type: Number}, nil
	case reflect.Bool:
		return Definition{Type: Boolean}, nil
	case reflect.Interface:
		// For interface{}/any, return empty schema (accepts anything)
		return Definition{}, nil
	default:
		return Definition{Type: String}, nil
	}
}

func generateStructSchema(t reflect.Type, seen map[reflect.Type]bool) (Definition, error) {
	seen[t] = true
	defer delete(seen, t)

	properties := make(map[string]Definition)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		isOmitempty := false
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			for _, part := range parts[1:] {
				if part == "omitempty" {
					isOmitempty = true
					break
				}
			}
		}

		// Generate schema for field type
		fieldSchema, err := generateSchema(field.Type, seen)
		if err != nil {
			return Definition{}, err
		}

		// Add description from tag
		if desc := field.Tag.Get("jsonschema_description"); desc != "" {
			fieldSchema.Description = desc
		}

		// Add enum values from tag
		if enumTag := field.Tag.Get("jsonschema_enum"); enumTag != "" {
			fieldSchema.Enum = strings.Split(enumTag, ",")
		}

		properties[fieldName] = fieldSchema

		// Fields without omitempty are required
		if !isOmitempty {
			required = append(required, fieldName)
		}
	}

	return Definition{
		Type:       Object,
		Properties: properties,
		Required:   required,
	}, nil
}

func generateArraySchema(t reflect.Type, seen map[reflect.Type]bool) (Definition, error) {
	elemSchema, err := generateSchema(t.Elem(), seen)
	if err != nil {
		return Definition{}, err
	}

	return Definition{
		Type:  Array,
		Items: &elemSchema,
	}, nil
}

func generateMapSchema(t reflect.Type, seen map[reflect.Type]bool) (Definition, error) {
	// Maps become objects with additionalProperties for the value type
	// For simplicity, we treat maps as objects (JSON only supports string keys)
	valueSchema, err := generateSchema(t.Elem(), seen)
	if err != nil {
		return Definition{}, err
	}

	// Note: JSON Schema supports additionalProperties as a schema, but our Definition
	// only supports bool. For maps, we return an object type.
	// Users needing full map support should use json.RawMessage.
	_ = valueSchema
	return Definition{
		Type:       Object,
		Properties: map[string]Definition{},
	}, nil
}

// Ptr is a helper function that returns a pointer to the given value.
// Useful for setting AdditionalProperties.
//
// Example:
//
//	schema.AdditionalProperties = jsonschema.Ptr(false)
func Ptr[T any](v T) *T {
	return &v
}
