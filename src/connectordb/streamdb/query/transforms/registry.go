package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

//TransformFunc is the function which transforms a given datapoint
type TransformFunc func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error)

//TransformGenerator is a function which "generates" the TransformFunc for the given transform
type TransformGenerator func(FunctionName string, Children ...TransformFunc) (TransformFunc, error)

//TransformArg represents an argument passed into the transform function
type TransformArg struct {
	Description string `json:"description"`       //A description of what the arg represents
	Optional    bool   `json:"optional"`          //Whether the arg is optional
	Default     string `json:"default,omitempty"` //If the arg is optional, what is its default value
	Constant    bool   `json:"constant"`          //If the argument must be a constant (ie, not part of a transform)
}

//Transform is the struct which holds the name, docstring, and generator for a transform function
type Transform struct {
	Name         string         `json:"name"`              //The name of the transform
	Description  string         `json:"description"`       //The description of the transform - a docstring
	InputSchema  string         `json:"ischema,omitempty"` //The schema of the input datapoint that the given transform expects (optional)
	OutputSchema string         `json:"oschema,omitempty"` //The schema of the output data that this transform gives (optional).
	Args         []TransformArg `json:"args"`              //The arguments that the transform accepts

	Generator TransformGenerator `json:"-"` //The generator function of the transform
}

var (
	//Registry is the map of all the transforms that are currently registered
	Registry = make(map[string]Transform)
)

//Register registers the transform with the system
func (t Transform) Register() error {
	if t.Name == "" || t.Generator == nil {
		err := fmt.Errorf("Attempted to register invalid transform: '%s'", t.Name)
		log.Error(err)
	}
	_, ok := Registry[t.Name]
	if ok {
		err := fmt.Errorf("A transform with the name '%s' already exists.", t.Name)
		log.Error("Transform registration failed: ", err)
		return err
	}

	Registry[t.Name] = t

	return nil
}

//Get returns the TransformFunc for the given name
func Get(name string, args ...TransformFunc) (TransformFunc, error) {
	t, ok := Registry[name]
	if !ok {
		return Err(fmt.Sprintf("Transform '%s' not found", name))
	}

	return t.Generator(name, args...)
}

//Err is the Error transform - a transform function that does nothing. It is a helper for when a transform func is to throw an error
func Err(errstring string) (TransformFunc, error) {
	err := errors.New(errstring)
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return nil, err
	}, err
}
