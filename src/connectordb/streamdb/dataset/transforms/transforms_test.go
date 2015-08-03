package transforms

import (
	"testing"

	"connectordb/streamdb/datastream"

	"github.com/connectordb/duck"
	"github.com/stretchr/testify/require"
)

type statelesstestcase struct {
	Args      []string
	Transform string
	Haserror  bool
	Haserror2 bool
	Input     datastream.Datapoint
	Output    *datastream.Datapoint
}

func statelessTransformTester(t *testing.T, testcases []statelesstestcase) {
	for _, c := range testcases {
		tr, err := Transforms[c.Transform](c.Args)
		if c.Haserror {
			require.Error(t, err, duck.JSONString(c))
		} else {
			require.NoError(t, err, duck.JSONString(c))
			dp, err := tr.Transform(&c.Input)
			if c.Haserror2 {
				require.Error(t, err, duck.JSONString(c))
			} else {
				require.NoError(t, err, duck.JSONString(c))
				if c.Output != nil {
					require.Equal(t, c.Output.String(), dp.String(), duck.JSONString(c))
				} else {
					require.Nil(t, dp, duck.JSONString(c))
				}
			}
		}
	}
}
