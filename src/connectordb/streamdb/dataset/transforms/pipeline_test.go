package transforms

// mark
import (
	. "connectordb/streamdb/datastream"
	"testing"

	"github.com/connectordb/duck"
	"github.com/stretchr/testify/require"
)

func TestPipeline(t *testing.T) {
	testcases := []struct {
		Pipeline  string
		Haserror  bool
		Haserror2 bool
		Input     *Datapoint
		Output    *Datapoint
	}{
		{"lt(5)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"has('test'):lt(1)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"ifhas('test'):lt(1)", false, false, &Datapoint{Data: 4}, nil},
		{"if(`has('test')`):lt(1)", false, false, &Datapoint{Data: 4}, nil},
		{"ifhas('test'):get('test'):lt(1)", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: false}},
		{"ifhas('tst'):get('test'):lt(1)", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, nil},
		{"ifhas('test'):get('test'):gt(1)", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: true}},
		{"ifhas('test", true, false, nil, nil},
		{"get('test')", false, true, &Datapoint{Data: 4}, nil},
	}

	for _, c := range testcases {
		tr, err := NewTransformPipeline(c.Pipeline)
		if c.Haserror {
			require.Error(t, err, duck.JSONString(c))
		} else {
			require.NoError(t, err, duck.JSONString(c))
			dp, err := tr.Transform(c.Input)
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
