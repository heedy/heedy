package transforms

import (
	. "connectordb/streamdb/datastream"
	"testing"

	"github.com/connectordb/duck"
	"github.com/stretchr/testify/require"
)

func TestPipelineGenerator(t *testing.T) {
	testcases := []struct {
		Pipeline       string
		HasSyntaxError bool
		Haserror2      bool
		Input          *Datapoint
		Output         *Datapoint
	}{
		{"true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"45.555", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 45.555}},
		{"\"string\"", false, false, &Datapoint{Data: 4}, &Datapoint{Data: "string"}},
		{"get()", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"4 < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"true or false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"false or false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"true and false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"true and (false or true)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"true and true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"true and not false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"if true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if true : 42", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 42}},
		{"if false", false, false, &Datapoint{Data: 4}, nil},
		{"if get() < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"get() > 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"get() > 3", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() >= 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() < 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"get() < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() <= 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() != 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"get() != 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() == 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"get() == 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"if get() < 5 and get() > 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if get() < 5 : if get() > 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if get() < 5 : if get() > 33", false, false, &Datapoint{Data: 4}, nil},
		{"if get() < 5 : get() > 33", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"has(\"test\"): get() < 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"if has(\"test\"): get() < 1", false, false, &Datapoint{Data: 4}, nil},
		{"if has(\"test\"):get(\"test\") < 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: false}},
		{"if has(\"tst\"):get(\"test\") < 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, nil},
		{"if has(\"test\"):get(\"test\") > 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: true}},
		{"if has(\"test\"", true, false, nil, nil},
		{"get(\"test\")", false, true, &Datapoint{Data: 4}, nil},

		//{"get() < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
	}

	for _, c := range testcases {

		result, err := ParseTransform(c.Pipeline)
		if c.HasSyntaxError {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err, duck.JSONString(c))

		dp, err := result(c.Input)
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

func TestParseTransform(t *testing.T) {
	// Valid pipeline
	{
		transform, err := ParseTransform("42")
		require.Nil(t, err)
		require.NotNil(t, transform)
	}

	// invalid pipeline
	{
		transform, err := ParseTransform("(")
		require.NotNil(t, err)
		require.Nil(t, transform)
	}
}
