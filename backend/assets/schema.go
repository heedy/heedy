package assets

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

// JSONSchema is used internally in assets to handle a version of json schema that pre-supposes
// that the root level is defining properties of an object
type JSONSchema struct {
	Schema map[string]interface{}

	s *gojsonschema.Schema
}

func NewSchema(schema map[string]interface{}) (*JSONSchema, error) {
	objectMap := make(map[string]interface{})
	objectMap["type"] = "object"
	objectMap["additionalProperties"] = false

	if v, ok := schema["type"]; ok {
		if v != "object" {
			return nil, errors.New("Schema must have type 'object'")
		}
		objectMap = schema
	} else {
		// Treat the schema as a prop map
		propMap := make(map[string]interface{})
		for k, v := range schema {
			switch k {
			// Allow these modifiers to go directly to the underlying schema object
			case "additionalProperties", "required":
				objectMap[k] = v
			default:
				propMap[k] = v
			}
		}
		objectMap["properties"] = propMap
	}

	s, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(objectMap))

	return &JSONSchema{
		Schema: objectMap,
		s:      s,
	}, err
}

// Validate ensures that the passed data conforms to the given schema
func (s *JSONSchema) Validate(data map[string]interface{}) error {
	res, err := s.s.Validate(gojsonschema.NewGoLoader(data))
	if err != nil {
		return err
	}
	if !res.Valid() {
		return errors.New(res.Errors()[0].String())
	}
	return nil
}

// ValidateWithDefaults both validates the given data, and inserts defaults for any missing
// values in the root object
func (s *JSONSchema) ValidateWithDefaults(data map[string]interface{}) (err error) {
	// The actual validation happens here
	defer func() {
		err = s.Validate(data)
	}()

	// Insert defaults into the object wherever the data is not provided
	propMapV, ok := s.Schema["properties"]
	if !ok {
		return
	}
	propMap, ok := propMapV.(map[string]interface{})
	if !ok {
		return
	}
	for k, v := range propMap {
		_, ok := data[k]
		if !ok {
			vmap, ok := v.(map[string]interface{})
			if ok {
				dval, ok := vmap["default"]
				if ok {
					data[k] = dval
				}
			}
		}
	}
	return
}
