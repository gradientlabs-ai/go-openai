package jsonschema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gradientlabs-ai/go-openai/jsonschema"
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

func TestGenerateSchemaForType(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Person{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if schema.Type != jsonschema.Object {
			t.Errorf("expected type Object, got %v", schema.Type)
		}
		if len(schema.Properties) != 2 {
			t.Errorf("expected 2 properties, got %d", len(schema.Properties))
		}
		if schema.Properties["name"].Type != jsonschema.String {
			t.Errorf("expected name to be String, got %v", schema.Properties["name"].Type)
		}
		if schema.Properties["age"].Type != jsonschema.Integer {
			t.Errorf("expected age to be Integer, got %v", schema.Properties["age"].Type)
		}
		if len(schema.Required) != 2 {
			t.Errorf("expected 2 required fields, got %d", len(schema.Required))
		}
	})

	t.Run("struct with descriptions", func(t *testing.T) {
		type Step struct {
			Explanation string `json:"explanation" jsonschema_description:"The reasoning for this step"`
			Output      string `json:"output" jsonschema_description:"The result of this step"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Step{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if schema.Properties["explanation"].Description != "The reasoning for this step" {
			t.Errorf("unexpected description: %v", schema.Properties["explanation"].Description)
		}
		if schema.Properties["output"].Description != "The result of this step" {
			t.Errorf("unexpected description: %v", schema.Properties["output"].Description)
		}
	})

	t.Run("struct with enum", func(t *testing.T) {
		type Status struct {
			State string `json:"state" jsonschema_enum:"pending,active,completed"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Status{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"pending", "active", "completed"}
		if len(schema.Properties["state"].Enum) != len(expected) {
			t.Errorf("expected %d enum values, got %d", len(expected), len(schema.Properties["state"].Enum))
		}
		for i, v := range expected {
			if schema.Properties["state"].Enum[i] != v {
				t.Errorf("expected enum[%d] to be %s, got %s", i, v, schema.Properties["state"].Enum[i])
			}
		}
	})

	t.Run("struct with omitempty (optional fields)", func(t *testing.T) {
		type Config struct {
			Name     string `json:"name"`
			Optional string `json:"optional,omitempty"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Config{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Only "name" should be required
		if len(schema.Required) != 1 {
			t.Errorf("expected 1 required field, got %d", len(schema.Required))
		}
		if schema.Required[0] != "name" {
			t.Errorf("expected 'name' to be required, got %v", schema.Required)
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

		schema, err := jsonschema.GenerateSchemaForType(Person{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		addressSchema := schema.Properties["address"]
		if addressSchema.Type != jsonschema.Object {
			t.Errorf("expected address type Object, got %v", addressSchema.Type)
		}
		if len(addressSchema.Properties) != 2 {
			t.Errorf("expected 2 address properties, got %d", len(addressSchema.Properties))
		}
	})

	t.Run("array field", func(t *testing.T) {
		type Response struct {
			Items []string `json:"items"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Response{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		itemsSchema := schema.Properties["items"]
		if itemsSchema.Type != jsonschema.Array {
			t.Errorf("expected items type Array, got %v", itemsSchema.Type)
		}
		if itemsSchema.Items == nil {
			t.Fatalf("expected items.Items to be set")
		}
		if itemsSchema.Items.Type != jsonschema.String {
			t.Errorf("expected items element type String, got %v", itemsSchema.Items.Type)
		}
	})

	t.Run("all primitive types", func(t *testing.T) {
		type AllTypes struct {
			String  string  `json:"string"`
			Int     int     `json:"int"`
			Int64   int64   `json:"int64"`
			Float64 float64 `json:"float64"`
			Bool    bool    `json:"bool"`
		}

		schema, err := jsonschema.GenerateSchemaForType(AllTypes{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if schema.Properties["string"].Type != jsonschema.String {
			t.Errorf("expected string type String")
		}
		if schema.Properties["int"].Type != jsonschema.Integer {
			t.Errorf("expected int type Integer")
		}
		if schema.Properties["int64"].Type != jsonschema.Integer {
			t.Errorf("expected int64 type Integer")
		}
		if schema.Properties["float64"].Type != jsonschema.Number {
			t.Errorf("expected float64 type Number")
		}
		if schema.Properties["bool"].Type != jsonschema.Boolean {
			t.Errorf("expected bool type Boolean")
		}
	})

	t.Run("skip json:- fields", func(t *testing.T) {
		type Secret struct {
			Public  string `json:"public"`
			Private string `json:"-"`
		}

		schema, err := jsonschema.GenerateSchemaForType(Secret{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(schema.Properties) != 1 {
			t.Errorf("expected 1 property, got %d", len(schema.Properties))
		}
		if _, exists := schema.Properties["Private"]; exists {
			t.Errorf("expected Private field to be skipped")
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type Data struct {
			Value string `json:"value"`
		}

		schema, err := jsonschema.GenerateSchemaForType(&Data{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if schema.Type != jsonschema.Object {
			t.Errorf("expected type Object, got %v", schema.Type)
		}
		if len(schema.Properties) != 1 {
			t.Errorf("expected 1 property, got %d", len(schema.Properties))
		}
	})
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
