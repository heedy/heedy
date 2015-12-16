/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package util

import "fmt"

func ExampleSchemaContainsObjectFields() {
	// false, type is a number
	result := SchemaContainsObjectFields(`{"type":"number"}`)
	fmt.Printf("%v\n", result)

	// true, parsing error (we take no chances)
	result = SchemaContainsObjectFields(`{"type}`)
	fmt.Printf("%v\n", result)

	// false, is an object, but no sub-objects
	result = SchemaContainsObjectFields(`{
		"type":"object",
		"properties":{
				"foo":{
					"type":"number"
				},
				"bar":{
					"type":"string"
				}
			}
		}`)
	fmt.Printf("%v\n", result)

	// true, is an object, and has sub-objects
	result = SchemaContainsObjectFields(`{
		"type":"object",
		"properties":{
				"foo":{
					"type":"object"
				}
			}
		}`)
	fmt.Printf("%v\n", result)
	// Output:
	// false
	// true
	// false
	// true
}
