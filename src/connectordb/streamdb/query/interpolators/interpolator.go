package interpolators

import (
	"connectordb/streamdb/datastream"
	"errors"
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//DefaultInterpolator is the one used when no interpolator is specified
var DefaultInterpolator = "closest"

var wordRegex *regexp.Regexp

// Converts an input into an array of tokens; strings will have quotes stripped
func tokenize(input string) []string {
	tokens := wordRegex.FindAllString(input, -1)

	// strip quotes
	for i, token := range tokens {
		if strings.HasSuffix(token, "\"") && strings.HasPrefix(token, "\"") && len(token) >= 2 {
			tokens[i] = token[1 : len(tokens)-1]
		}
	}

	return tokens
}

//Interpolator is an interface which given a timestamp, returns the appropriate
//datapoint. Interpolator is guaranteed to be called with increasing or equal timestamps
//since the dataset is constructed iteratively. "registered" interpolators are given
//a DataRange and a string array of arguments passed by the user
type Interpolator interface {
	Interpolate(timestamp float64) (tdp *datastream.Datapoint, err error)
	Close()
}

//InterpolatorGenerator is the signature of a function that generates an interpolator
type InterpolatorGenerator func(dr datastream.DataRange, args []string) (Interpolator, error)

//InterpolatorArg represents an argument to the interpolator
type InterpolatorArg struct {
	Description string `json:"description"`       //A description of what the arg represents
	Optional    bool   `json:"optional"`          //Whether the arg is optional
	Default     string `json:"default,omitempty"` //If the arg is optional, what is its default value

}

//InterpolatorDescription describes an interpolator - it functions as the documentation
type InterpolatorDescription struct {
	Name         string            `json:"name"`              //The name of the interpolator
	Description  string            `json:"description"`       //The description of the transform - a docstring
	InputSchema  string            `json:"ischema,omitempty"` //The schema of the input datapoints that the given interpolator expects (optional)
	OutputSchema string            `json:"oschema,omitempty"` //The schema of the output data that this interpolator gives (optional).
	Args         []InterpolatorArg `json:"args"`              //The arguments that the interpolator accepts

	Generator InterpolatorGenerator `json:"-"` //The generator function of the transform
}

//Interpolators is the map of all registered interpolations
var InterpolatorRegistry = make(map[string]InterpolatorDescription)

func (i InterpolatorDescription) Register() error {
	if i.Name == "" || i.Generator == nil {
		err := fmt.Errorf("Attempted to register invalid interpolator: '%s'", i.Name)
		log.Error(err)
	}
	_, ok := InterpolatorRegistry[i.Name]
	if ok {
		err := fmt.Errorf("An interpolator with the name '%s' already exists.", i.Name)
		log.Error("Interpolator registration failed: ", err)
		return err
	}

	InterpolatorRegistry[i.Name] = i

	return nil
}

//Get gets an interpolator given the string which defines the interpolator and all arguments
//it takes
func Get(dr datastream.DataRange, interp string) (Interpolator, error) {
	if interp == "" {
		interp = DefaultInterpolator
	}

	tokens := tokenize(interp)
	if len(tokens) == 0 {
		return nil, errors.New("no interpolator found")
	}

	interpolatorName := tokens[0]

	idesc, ok := InterpolatorRegistry[interpolatorName]
	if !ok {
		return nil, errors.New("Could not find '" + interpolatorName + "' interpolator.")
	}

	args := []string{}
	if len(tokens) > 1 {
		args = tokens[1:]
	}

	return idesc.Generator(dr, args)

}

func init() {
	re, err := regexp.Compile("\".+?\"|\\w+")

	if err != nil {
		panic(err.Error())
	}

	wordRegex = re

	//register the builtin interpolators
	before.Register()
	after.Register()
	closest.Register()
}
