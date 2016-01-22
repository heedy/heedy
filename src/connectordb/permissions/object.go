package permissions

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"
)

/*
// ReadObject is given a pointer to the given struct
func ReadObject(prefix string, access *config.AccessLevel, obj interface{}) error {
	amap := access.GetMap()

	otype := reflect.TypeOf(obj)
	otype = otype.Elem()
	oval := reflect.ValueOf(obj)
	oval = val.Elem() // Follow the pointer
	for i := 0; i < aval.NumField(); i++ {
		t := otype.Field(i)
		v := oval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			// The value is relevant - check if we are allowed to read it
			if !amap[prefix+tags[0]] {
				// We are NOT allowed to read it. Clear the value
				// NOTE: Only the types that are actually present in User/Device/Stream are available here
				switch v.Kind() {
				case reflect.String:
					v.SetString("")
				case reflect.Bool:
					v.SetBool(false)
				case reflect.Int64, reflect.Int:
					v.SetInt(0)
				}
			}
		}
	}

	return nil
}
*/

// ReadObjectToMap is given a pointer to the given struct
func ReadObjectToMap(prefix string, amap map[string]bool, obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	oval := reflect.ValueOf(obj)
	oval = oval.Elem()
	otype := oval.Type()
	for i := 0; i < oval.NumField(); i++ {
		t := otype.Field(i)
		v := oval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			haspermission, ok := amap[prefix+tags[0]]
			if !ok {
				log.Fatalf("Could not find field '%s' in access level! This is an error in ConnectorDB code. You must have changed the users/devices/streams structs without updating access levels.", prefix+tags[0])
			}
			// The value is relevant - check if we are allowed to read it
			if haspermission {
				// We are allowed to read it. Write the interface to the map
				result[tags[0]] = v.Interface()
			}
		}
	}

	return result
}

/*
func WriteObject(prefix string, access *config.AccessLevel, original interface{}, modified interface{}) error {
	amap := access.GetMap()

	oval := reflect.ValueOf(*original)
	mval := reflect.ValueOf(*modified)

	otype := reflect.TypeOf(*original)

	for i := 0; i < oval.NumField(); i++ {
		t := otype.Field(i)
		ov := oval.Field(i)
		mv := mval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			tag := prefix + tags[0]
			switch mv.Kind() {
			case reflect.String:
				if mv.String() != ov.String() {
					if !amap[tag] {
						return fmt.Errorf("This device does not have permissions to write %s", tag)
					}
				}
			case reflect.Bool:
				if mv.Bool() != ov.Bool() {
					if !amap[tag] {
						return fmt.Errorf("This device does not have permissions to write %s", tag)
					}
				}
			case reflect.Int64:
				if mv.Int() != ov.Int() {
					if !amap[tag] {
						return fmt.Errorf("This device does not have permissions to write %s", tag)
					}
				}
			}

		}
	}
}
*/

// WriteMap takes the data given in the data map, and writes it to original
func WriteObjectFromMap(prefix string, amap map[string]bool, obj interface{}, data map[string]interface{}) error {

	// First make sure that we have permissions to write all the values that we are to modify
	for key := range data {
		hasAccess, ok := amap[prefix+key]
		if !ok {
			return fmt.Errorf("Unrecognized field '%s'", key)
		}
		if !hasAccess {
			return fmt.Errorf("This device does not have permissions to write %s", key)
		}
	}

	// So we have write access to all of the given fields - let's write them!
	oval := reflect.ValueOf(obj)
	oval = oval.Elem()
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
			}

		}
	}

	return nil
}
