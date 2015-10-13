package transforms

import (
	"connectordb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

// A StatusFlag holds some state that the interpreter thinks is important.
type StatusFlag int

const (
	// This flag is naturally set, nothing important.
	NilFlag StatusFlag = iota
	// The flag sent when this is the last datapoint in the system.
	LastDatapoint StatusFlag = iota
	// This flag is sent to ask if the function returns a constant (string/number)
	constantCheck StatusFlag = iota
	// this flag is set as a reply if the "constantCheck" passes
	constantCheckTrue StatusFlag = iota
	// Flag sent through to indicate the "last" datapoint was dropped
	lastDatapointDropped = iota
)

var (
	ErrNotFloat  = errors.New("Value is not a float")
	ErrNotString = errors.New("Value is not a string")
	ErrNotBool   = errors.New("Value is not a bool")
)

// NewTransformEnvironment creates an environment ready to process the data
func NewTransformEnvironment(datapoint *datastream.Datapoint) *TransformEnvironment {
	return &TransformEnvironment{NilFlag, datapoint, nil}
}

// TransformEnvironment represents the current state of the interpreter.
type TransformEnvironment struct {
	// Flag holds some bit of interpreter status
	Flag StatusFlag

	// The data to be processed
	Datapoint *datastream.Datapoint

	// An error the system has encountered, set this on a failure
	Error error
}

// Produces a non-deep copy of this environment
func (t *TransformEnvironment) Copy() *TransformEnvironment {
	if t == nil {
		return &TransformEnvironment{Datapoint: nil}
	}

	n := TransformEnvironment{t.Flag, t.Datapoint.Copy(), t.Error}
	return &n
}

// CanProcess checks that the environment is not nil, the data is not nil,
// and the error is not set; in other words, nothing has gone horribly wrong with
// the interpreter environment.
func (t *TransformEnvironment) CanProcess() bool {
	if t == nil || t.Datapoint == nil || t.Error != nil {
		return false
	}

	return true
}

// SetError sets an error code and returns the TransformEnvironment
func (t *TransformEnvironment) SetError(err error) *TransformEnvironment {
	if t == nil {
		return &TransformEnvironment{Error: err}
	}

	if err == nil {
		return t
	}

	t.Error = err
	return t
}

// SetErrorString creates a new error from the given string and sets it
func (t *TransformEnvironment) SetErrorString(errstr string) *TransformEnvironment {
	err := errors.New(errstr)
	return t.SetError(err)
}

// Applies a transform to this datapoint and returns the result
func (t *TransformEnvironment) Apply(transform TransformFunc) *TransformEnvironment {
	return transform(t)
}

// Gets the value of the datapoint as a float if possible, if not possible
// sets the error to be a conversion error
func (t *TransformEnvironment) GetFloat() (value float64, ok bool) {
	if !t.CanProcess() {
		t.SetError(ErrNotFloat)
		return 0, false
	}

	v, ok := duck.Float(t.Datapoint.Data)
	if !ok {
		t.SetError(ErrNotFloat)
	}
	return v, ok
}

// Gets the value of the datapoint as a bool if possible, if not possible
// sets the error to be a conversion error
func (t *TransformEnvironment) GetBool() (value bool, ok bool) {
	if !t.CanProcess() {
		t.SetError(ErrNotBool)
		return false, false
	}

	v, ok := duck.Bool(t.Datapoint.Data)
	if !ok {
		t.SetError(ErrNotBool)
	}
	return v, ok
}

// Gets the value of the datapoint as a bool if possible, if not possible
// sets the error to be a conversion error
func (t *TransformEnvironment) GetString() (value string, ok bool) {
	if !t.CanProcess() {
		t.SetError(ErrNotString)
		return "", false
	}

	v, ok := duck.String(t.Datapoint.String)
	if !ok {
		t.SetError(ErrNotBool)
	}
	return v, ok
}

// Sets the data for the datpaoint
func (t *TransformEnvironment) SetData(value interface{}) *TransformEnvironment {
	if t == nil {
		return &TransformEnvironment{Error: errors.New("Nil Environment Error")}
	}

	t.Datapoint.Data = value
	return t
}

// Sets the data for the datpaoint
func (t *TransformEnvironment) SetFlag(value StatusFlag) *TransformEnvironment {
	if t == nil {
		return &TransformEnvironment{Flag: value}
	}

	t.Flag = value
	return t
}
