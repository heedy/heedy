// Package users provides an API for managing user information.
package users

import (
	"errors"
	"reflect"
	//"fmt"
)

// A PermissionLevel within the system. These determine which devices can
// edit/read all the data in the database.
type PermissionLevel uint

const (
	// NOBODY is the highest permission level, nobody has it; if something needs
	// to be changed with a NOBODY permission, it must be done using SQL. These
	// operations are to keep dangerous things from happening, like modifying
	// primary keys
	NOBODY = PermissionLevel(6)
	// ROOT is the highest permission level given to admin devices/users
	ROOT = PermissionLevel(5)
	// USER is the permission level given to devices that can modify user data
	// and act on a user's behalf
	USER = PermissionLevel(4)
	// DEVICE the device that owns a stream or can operate on itself.
	DEVICE = PermissionLevel(3)
	// FAMILY is for devices with the same owner, but no edit permissions
	FAMILY = PermissionLevel(2)
	// ENABLED is for any device that can do reading in the system
	ENABLED = PermissionLevel(1)
	// ANYBODY is for doing completely unpriviliged operations.
	ANYBODY = PermissionLevel(0)
)

func strToPermissionLevel(s string) (PermissionLevel, error) {
	switch s {
	case "nobody":
		return NOBODY, nil
	case "root":
		return ROOT, nil
	case "user":
		return USER, nil
	case "device":
		return DEVICE, nil
	case "family":
		return FAMILY, nil
	case "enabled":
		return ENABLED, nil
	case "anybody":
		return ANYBODY, nil
	}

	return ANYBODY, errors.New("Given string is not a valid permission type")
}

// Gte checks that the given permission is at least what the desired one should be
func (actual PermissionLevel) Gte(desired PermissionLevel) bool {
	return uint(actual) >= uint(desired)
}

func revertUneditableFields(toChange reflect.Value, originalValue reflect.Value, p PermissionLevel) int {

	//fmt.Printf("Getting original elem %v\n", originalValue.Kind())
	originalValueReflect := originalValue //.Elem()

	//fmt.Println("done getting elem")
	changeNumber := 0
	for i := 0; i < originalValueReflect.NumField(); i++ {
		// Grab the fields for reflection
		originalValueField := originalValueReflect.Field(i)
		typeField := originalValueReflect.Type().Field(i)

		// Check what kind of modifiable permission we need to edit
		modifiable := typeField.Tag.Get("modifiable")

		// By default, we don't allow modification
		if modifiable == "" {
			modifiable = "nobody"
		}

		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, originalValueField.Interface(), modifiable)

		//fmt.Printf("Field name: %v, modifiable %v\n",originalTypeField.Name, originalValueField.String(),  modifiable)

		// If we don't have enough permissions, reset the field from original
		requiredPermissionsForField, _ := strToPermissionLevel(modifiable)
		if !p.Gte(requiredPermissionsForField) {
			//fmt.Printf("Setting field\n")
			if !reflect.DeepEqual(toChange.Elem().Field(i).Interface(), originalValueField.Interface()) {
				toChange.Elem().Field(i).Set(originalValueField)
				changeNumber++
			}
		}
	}

	// and bob's your uncle!
	return changeNumber
}
