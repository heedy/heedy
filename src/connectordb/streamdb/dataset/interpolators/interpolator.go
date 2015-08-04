package interpolators

import (
	"connectordb/streamdb/datastream"
	"errors"
	"regexp"
	"strings"
)

//DefaultInterpolator is the one used when no interpolator is specified
var DefaultInterpolator = "closest"

var wordRegex *regexp.Regexp

func init() {
	re, err := regexp.Compile("\".+?\"|\\w+")

	if err != nil {
		panic(err.Error())
	}

	wordRegex = re
}

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

//Interpolators is the map of all registered interpolations
var Interpolators = map[string]InterpolatorGenerator{
	"before":  NewBeforeInterpolator,
	"after":   NewAfterInterpolator,
	"closest": NewClosestInterpolator,
}

//GetInterpolator gets an interpolator given the string which defines the interpolator and all arguments
//it takes
func GetInterpolator(dr datastream.DataRange, interp string) (Interpolator, error) {
	if interp == "" {
		interp = DefaultInterpolator
	}

	tokens := tokenize(interp)
	if len(tokens) == 0 {
		return nil, errors.New("no interpolater found")
	}

	interpolatorName := tokens[0]

	ifunc, ok := Interpolators[interpolatorName]
	if !ok {
		return nil, errors.New("Could not find '" + interpolatorName + "' interpolator.")
	}

	args := []string{}
	if len(tokens) > 1 {
		args = tokens[1:]
	}

	return ifunc(dr, args)

}
