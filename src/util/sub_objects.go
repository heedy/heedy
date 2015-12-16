/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package util

import "github.com/xeipuuv/gojsonschema"

const (
	// I'm really sorry about this
	validNonObject = `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "additionalProperties": false,
  "properties": {
    "properties": {
      "patternProperties": {
        ".+": {
          "properties": {
            "type": {
              "enum": [
                "boolean",
                "integer",
                "number",
                "string"
              ]
            },
            "$ref": {
              "type": "string",
              "pattern": "^connectordb_definitions.json#/"
            }
          },
          "type": "object"
        }
      },
      "type": "object"
    },
    "type": {
      "type": "string"
    }
  },
  "type": "object"
}`
)

// Returns true if the given jsonschema contains object fields
// returns true on a parsing error
func SchemaContainsObjectFields(jsonSchema string) bool {
	schemaLoader := gojsonschema.NewStringLoader(validNonObject)
	documentLoader := gojsonschema.NewStringLoader(jsonSchema)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return true
	}

	return !result.Valid()
}
