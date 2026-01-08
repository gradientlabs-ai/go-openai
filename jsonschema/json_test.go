package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/gradientlabs-ai/go-openai/jsonschema"
	invopopjsonschema "github.com/invopop/jsonschema"
)

func TestDefinition_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		def  jsonschema.Definition
		want string
	}{
		{
			name: "Test with empty Definition",
			def:  jsonschema.Definition{},
			want: `{"properties":{}}`,
		},
		{
			name: "Test with Definition properties set",
			def: jsonschema.Definition{
				Type:        jsonschema.String,
				Description: "A string type",
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"string",
   "description":"A string type",
   "properties":{
      "name":{
         "type":"string",
         "properties":{}
      }
   }
}`,
		},
		{
			name: "Test with nested Definition properties",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string",
               "properties":{}
            },
            "age":{
               "type":"integer",
               "properties":{}
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with complex nested Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"user": {
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"name": {
								Type: jsonschema.String,
							},
							"age": {
								Type: jsonschema.Integer,
							},
							"address": {
								Type: jsonschema.Object,
								Properties: map[string]jsonschema.Definition{
									"city": {
										Type: jsonschema.String,
									},
									"country": {
										Type: jsonschema.String,
									},
								},
							},
						},
					},
				},
			},
			want: `{
   "type":"object",
   "properties":{
      "user":{
         "type":"object",
         "properties":{
            "name":{
               "type":"string",
               "properties":{}
            },
            "age":{
               "type":"integer",
               "properties":{}
            },
            "address":{
               "type":"object",
               "properties":{
                  "city":{
                     "type":"string",
                     "properties":{}
                  },
                  "country":{
                     "type":"string",
                     "properties":{}
                  }
               }
            }
         }
      }
   }
}`,
		},
		{
			name: "Test with Array type Definition",
			def: jsonschema.Definition{
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
				Properties: map[string]jsonschema.Definition{
					"name": {
						Type: jsonschema.String,
					},
				},
			},
			want: `{
   "type":"array",
   "items":{
      "type":"string",
      "properties":{

      }
   },
   "properties":{
      "name":{
         "type":"string",
         "properties":{}
      }
   }
}`,
		},
		{
			name: "Test with AdditionalProperties false (strict mode)",
			def: func() jsonschema.Definition {
				f := false
				return jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"name": {
							Type: jsonschema.String,
						},
					},
					Required:             []string{"name"},
					AdditionalProperties: &f,
				}
			}(),
			want: `{
   "type":"object",
   "properties":{
      "name":{
         "type":"string",
         "properties":{}
      }
   },
   "required":["name"],
   "additionalProperties":false
}`,
		},
		{
			name: "Test with AdditionalProperties true",
			def: func() jsonschema.Definition {
				t := true
				return jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"data": {
							Type: jsonschema.String,
						},
					},
					AdditionalProperties: &t,
				}
			}(),
			want: `{
   "type":"object",
   "properties":{
      "data":{
         "type":"string",
         "properties":{}
      }
   },
   "additionalProperties":true
}`,
		},
		{
			name: "Test with AdditionalProperties nil (omitted)",
			def: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"value": {
						Type: jsonschema.Number,
					},
				},
				AdditionalProperties: nil,
			},
			want: `{
   "type":"object",
   "properties":{
      "value":{
         "type":"number",
         "properties":{}
      }
   }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantBytes := []byte(tt.want)
			var want map[string]interface{}
			err := json.Unmarshal(wantBytes, &want)
			if err != nil {
				t.Errorf("Failed to Unmarshal JSON: error = %v", err)
				return
			}

			got := structToMap(t, tt.def)
			gotPtr := structToMap(t, &tt.def)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, want)
			}
			if !reflect.DeepEqual(gotPtr, want) {
				t.Errorf("MarshalJSON() gotPtr = %v, want %v", gotPtr, want)
			}
		})
	}
}

func structToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	gotBytes, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Failed to Marshal JSON: error = %v", err)
		return nil
	}

	var got map[string]any
	err = json.Unmarshal(gotBytes, &got)
	if err != nil {
		t.Errorf("Failed to Unmarshal JSON: error =  %v", err)
		return nil
	}
	return got
}

func TestPtr(t *testing.T) {
	f := jsonschema.Ptr(false)
	if *f != false {
		t.Errorf("expected false, got %v", *f)
	}

	tr := jsonschema.Ptr(true)
	if *tr != true {
		t.Errorf("expected true, got %v", *tr)
	}

	s := jsonschema.Ptr("hello")
	if *s != "hello" {
		t.Errorf("expected 'hello', got %v", *s)
	}
}

func TestGenerateSchema(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		schema := jsonschema.GenerateSchema[Person]()

		if schema.Type != "object" {
			t.Errorf("expected type object, got %v", schema.Type)
		}
		if schema.Properties.Len() != 2 {
			t.Errorf("expected 2 properties, got %d", schema.Properties.Len())
		}
		if schema.Properties.GetPair("name").Value.Type != "string" {
			t.Errorf("expected name to be string")
		}
		if schema.Properties.GetPair("age").Value.Type != "integer" {
			t.Errorf("expected age to be integer")
		}
		// AdditionalProperties should be false (strict mode)
		if schema.AdditionalProperties != invopopjsonschema.FalseSchema {
			t.Errorf("expected AdditionalProperties to be FalseSchema")
		}
	})

	t.Run("struct with descriptions", func(t *testing.T) {
		type Step struct {
			Explanation string `json:"explanation" jsonschema_description:"The reasoning for this step"`
			Output      string `json:"output" jsonschema_description:"The result of this step"`
		}

		schema := jsonschema.GenerateSchema[Step]()

		explanationProp := schema.Properties.GetPair("explanation")
		if explanationProp.Value.Description != "The reasoning for this step" {
			t.Errorf("unexpected description: %v", explanationProp.Value.Description)
		}
		outputProp := schema.Properties.GetPair("output")
		if outputProp.Value.Description != "The result of this step" {
			t.Errorf("unexpected description: %v", outputProp.Value.Description)
		}
	})

	t.Run("struct with enum", func(t *testing.T) {
		type Status struct {
			State string `json:"state" jsonschema:"enum=pending,enum=active,enum=completed"`
		}

		schema := jsonschema.GenerateSchema[Status]()

		stateProp := schema.Properties.GetPair("state")
		expected := []any{"pending", "active", "completed"}
		if len(stateProp.Value.Enum) != len(expected) {
			t.Errorf("expected %d enum values, got %d", len(expected), len(stateProp.Value.Enum))
		}
	})

	t.Run("nested struct", func(t *testing.T) {
		type Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
		}
		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		schema := jsonschema.GenerateSchema[Person]()

		addressProp := schema.Properties.GetPair("address")
		if addressProp.Value.Type != "object" {
			t.Errorf("expected address type object, got %v", addressProp.Value.Type)
		}
		if addressProp.Value.Properties.Len() != 2 {
			t.Errorf("expected 2 address properties, got %d", addressProp.Value.Properties.Len())
		}
		// Nested object should also have AdditionalProperties false
		if addressProp.Value.AdditionalProperties != invopopjsonschema.FalseSchema {
			t.Errorf("expected nested AdditionalProperties to be FalseSchema")
		}
	})

	t.Run("array field", func(t *testing.T) {
		type Response struct {
			Items []string `json:"items"`
		}

		schema := jsonschema.GenerateSchema[Response]()

		itemsProp := schema.Properties.GetPair("items")
		if itemsProp.Value.Type != "array" {
			t.Errorf("expected items type array, got %v", itemsProp.Value.Type)
		}
		if itemsProp.Value.Items == nil {
			t.Fatalf("expected items.Items to be set")
		}
		if itemsProp.Value.Items.Type != "string" {
			t.Errorf("expected items element type string, got %v", itemsProp.Value.Items.Type)
		}
	})

	t.Run("embedded struct", func(t *testing.T) {
		type Base struct {
			ID string `json:"id"`
		}
		type Extended struct {
			Base
			Name string `json:"name"`
		}

		schema := jsonschema.GenerateSchema[Extended]()

		// Embedded struct fields should be flattened
		if schema.Properties.GetPair("id") == nil {
			t.Errorf("expected embedded 'id' field to be present")
		}
		if schema.Properties.GetPair("name") == nil {
			t.Errorf("expected 'name' field to be present")
		}
	})

	t.Run("time.Time field", func(t *testing.T) {
		type Event struct {
			Name      string    `json:"name"`
			Timestamp time.Time `json:"timestamp"`
		}

		schema := jsonschema.GenerateSchema[Event]()

		timestampProp := schema.Properties.GetPair("timestamp")
		// time.Time should be represented as a string with date-time format
		if timestampProp.Value.Type != "string" {
			t.Errorf("expected timestamp type string, got %v", timestampProp.Value.Type)
		}
		if timestampProp.Value.Format != "date-time" {
			t.Errorf("expected timestamp format date-time, got %v", timestampProp.Value.Format)
		}
	})

	t.Run("pointer field", func(t *testing.T) {
		type Data struct {
			Value  string  `json:"value"`
			OptPtr *string `json:"opt_ptr,omitempty"`
		}

		schema := jsonschema.GenerateSchema[Data]()

		if schema.Type != "object" {
			t.Errorf("expected type object, got %v", schema.Type)
		}
		// Both fields should exist
		if schema.Properties.GetPair("value") == nil {
			t.Errorf("expected 'value' field to be present")
		}
		if schema.Properties.GetPair("opt_ptr") == nil {
			t.Errorf("expected 'opt_ptr' field to be present")
		}
	})
}
