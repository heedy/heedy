/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package functions

import (
	"connectordb/datastream"
	"connectordb/query/transforms"
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
	Args        []transforms.TransformFunc
	HasError    bool
	Description string

	Tests []TestCaseElement
}

func (tc TestCase) Run(t *testing.T) {
	tf, err := transforms.InstantiateRegisteredFunction(tc.Name, tc.Args...)
	if tc.HasError {
		require.Error(t, err, fmt.Sprintf("%v", tc))
		return
	}
	require.NoError(t, err, fmt.Sprintf("%v", tc))

	for _, test := range tc.Tests {
		environment := transforms.NewTransformEnvironment(test.Input)
		tmp := tf(environment)
		out := tmp.Datapoint
		err := tmp.Error

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
