package functions

import (
	"connectordb/streamdb/datastream"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCaseElement struct {
	Input       *datastream.Datapoint
	Output      *datastream.Datapoint
	HasError    bool
	Description string
}

type TestCase struct {
	Name        string
	Args        []TransformFunc
	HasError    bool
	Description string

	Tests []TestCaseElement
}

func ConstTransform(c interface{}) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return &datastream.Datapoint{Data: c}, nil
	}
}

func (tc TestCase) Run(t *testing.T) {
	tf, err := Get(tc.Name, tc.Args...)
	if tc.HasError {
		require.Error(t, err, fmt.Sprintf("%v", tc))
		return
	}
	require.NoError(t, err, fmt.Sprintf("%v", tc))

	for _, test := range tc.Tests {
		out, err := tf(test.Input)
		if test.HasError {
			require.Error(t, err, fmt.Sprintf("%v", test))
			return
		}
		require.NoError(t, err, fmt.Sprintf("%v", test))
		if test.Output == nil {
			require.Nil(t, out)
		} else {
			if !out.IsEqual(*test.Output) {
				require.Equal(t, test.Output.String(), out.String(), fmt.Sprintf("%v", test))
			}
		}

	}
}
