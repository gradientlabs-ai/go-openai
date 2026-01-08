// Package jsonschema provides functionality for generating and representing JSON schemas.
// It can be used with OpenAI's chat completion "function call" and "structured outputs" features.
package jsonschema

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
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
// For generating schemas from Go types, use GenerateSchema[T]() instead.
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

// GenerateSchema generates a JSON schema from a Go type using reflection.
// It uses github.com/invopop/jsonschema with settings optimized for OpenAI structured outputs:
//   - AdditionalProperties is always false (strict mode)
//   - No $ref usage (schemas are fully inlined)
//
// Supported struct tags:
//   - `json:"field_name"` for the JSON field name
//   - `jsonschema:"title=X,description=Y,enum=a,enum=b"` for schema properties
//   - `jsonschema_description:"..."` for field descriptions
//
// Example:
//
//	type Step struct {
//	    Explanation string `json:"explanation" jsonschema_description:"The reasoning for this step"`
//	    Output      string `json:"output" jsonschema_description:"The result of this step"`
//	}
//
//	type MathResponse struct {
//	    Steps       []Step `json:"steps" jsonschema_description:"List of solution steps"`
//	    FinalAnswer string `json:"final_answer" jsonschema_description:"The final answer"`
//	}
//
//	schema := jsonschema.GenerateSchema[MathResponse]()
func GenerateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

