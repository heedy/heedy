package connectordb

import (
	"fmt"
	"reflect"
	"strings"
)

// WriteObjectFromMap takes a pointer to a go struct, which has json fields defined,
// and writes the changes given in the data map passed in. The passed in object is
// therefore modified to reflect the changes.
func WriteObjectFromMap(obj interface{}, data map[string]interface{}) error {

	oval := reflect.ValueOf(obj)
	oval = oval.Elem() // oval is the pointer
	otype := oval.Type()

	for i := 0; i < oval.NumField(); i++ {
		t := otype.Field(i)
		ov := oval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			iface, ok := data[tags[0]]
			if ok {
				// The data to modify this tag is available!

				switch ov.Kind() {
				case reflect.String:
					v, ok := iface.(string)
					if !ok {
						return fmt.Errorf("Value of field '%s' must be a string", tags[0])
					}
					ov.SetString(v)
				case reflect.Bool:
					v, ok := iface.(bool)
					if !ok {
						return fmt.Errorf("Value of field '%s' must be a boolean", tags[0])
					}
					ov.SetBool(v)
				case reflect.Int64:
					// JSON unmarshals numbers as float64
					// https://golang.org/pkg/encoding/json/#Unmarshal
					v, ok := iface.(float64)
					var vint int64
					if !ok {
						// This field might be modified by a ConnectorDB internal program, so check if it is an integer
						mv := reflect.ValueOf(iface)
						k := mv.Kind()
						if k == reflect.Int || k == reflect.Int64 {
							vint = mv.Int()
						} else {
							return fmt.Errorf("Value of field '%s' must be a number", tags[0])
						}

					} else {
						vint = int64(v)
						if float64(vint) != v {
							return fmt.Errorf("Value of field '%s' must be an integer", tags[0])
						}
					}

					ov.SetInt(vint)
				}

				// Now delete the element from the map, so we can detect extra fields
				// TODO: How to do this without modifying the map?
				delete(data, tags[0])
			}

		}
	}

	if len(data) > 0 {
		for key := range data {
			// There has to be a better way to get a key
			return fmt.Errorf("Field '%s' not found", data[key])
		}
	}

	return nil
}
