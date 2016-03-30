package permissions

import (
	pconfig "config/permissions"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// ErrNoAccess is returned when no access at all is given to the resource
var ErrNoAccess = errors.New("Can't access this resource.")

func CheckIfUpdateFieldsPermitted(perm *pconfig.Permissions, useraccess *pconfig.AccessLevel, deviceaccess *pconfig.AccessLevel, prefix string, updates map[string]interface{}) error {
	ua := GetWriteAccess(perm, useraccess).GetMap()
	da := GetWriteAccess(perm, deviceaccess).GetMap()

	p, ok := ua["can_access_"+prefix]
	if !ok {
		log.Fatalf("Could not find field 'can_access_%s' in rwaccess! This is an error in ConnectorDB code.", prefix)
	}
	p2 := da["can_access_"+prefix]
	if !p || !p2 {
		return ErrNoAccess
	}

	prefix = prefix + "_"

	for key := range updates {
		hasAccess1, ok := ua[prefix+key]
		if !ok {
			return fmt.Errorf("Unrecognized field '%s'", key)
		}
		hasAccess2, ok := da[prefix+key]
		if !hasAccess1 || !hasAccess2 {
			return fmt.Errorf("This device does not have permissions to write %s", key)
		}
	}

	return nil
}

// DeleteDisallowedFields sets the fields that are not permitted to their corresponding null values.
// Note that booleans are set to false, which makes it impossible to tell if a boolean's actual
// value is false, or if one simply does not have permissions to read the value
func DeleteDisallowedFields(perm *pconfig.Permissions, useraccess *pconfig.AccessLevel, deviceaccess *pconfig.AccessLevel, prefix string, obj interface{}) error {
	ua := GetReadAccess(perm, useraccess).GetMap()
	da := GetReadAccess(perm, deviceaccess).GetMap()

	p, ok := ua["can_access_"+prefix]
	if !ok {
		log.Fatalf("Could not find field 'can_access_%s' in rwaccess! This is an error in ConnectorDB code.", prefix)
	}
	p2 := da["can_access_"+prefix]
	if !p || !p2 {
		// Return nil map if no access
		return ErrNoAccess
	}

	prefix = prefix + "_"

	oval := reflect.ValueOf(obj)
	oval = oval.Elem()
	otype := oval.Type()
	for i := 0; i < oval.NumField(); i++ {
		t := otype.Field(i)
		v := oval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			p, ok = ua[prefix+tags[0]]
			if !ok {
				log.Fatalf("Could not find field '%s' in access level! This is an error in ConnectorDB code. You must have changed the users/devices/streams structs without updating access levels.", prefix+tags[0])
			}
			p2 = da[prefix+tags[0]]
			// The value is relevant - check if we are allowed to read it
			if !p || !p2 {
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

// ReadObjectToMap returns the map of the data. If this user/device can't access the object, ErrNoAccess is returned
func ReadObjectToMap(perm *pconfig.Permissions, useraccess *pconfig.AccessLevel, deviceaccess *pconfig.AccessLevel, prefix string, obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	ua := GetReadAccess(perm, useraccess).GetMap()
	da := GetReadAccess(perm, deviceaccess).GetMap()

	p, ok := ua["can_access_"+prefix]
	if !ok {
		log.Fatalf("Could not find field 'can_access_%s' in rwaccess! This is an error in ConnectorDB code.", prefix)
	}
	p2 := da["can_access_"+prefix]
	if !p || !p2 {
		// Return nil map if no access
		return nil, ErrNoAccess
	}

	prefix = prefix + "_"

	oval := reflect.ValueOf(obj)
	oval = oval.Elem()
	otype := oval.Type()
	for i := 0; i < oval.NumField(); i++ {
		t := otype.Field(i)
		v := oval.Field(i)

		// The json tag can have commas in it (like omitempty)
		tags := strings.Split(t.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" {
			p, ok = ua[prefix+tags[0]]
			if !ok {
				log.Fatalf("Could not find field '%s' in access level! This is an error in ConnectorDB code. You must have changed the users/devices/streams structs without updating access levels.", prefix+tags[0])
			}
			p2 = da[prefix+tags[0]]
			// The value is relevant - check if we are allowed to read it
			if p && p2 {
				// We are allowed to read it. Write the interface to the map
				result[tags[0]] = v.Interface()
			}
		}
	}

	return result, nil
}
