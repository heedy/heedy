package transforms

import (
	. "connectordb/streamdb/datastream"
	"errors"
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
		// Identity functions
		{"true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"45.555", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 45.555}},
		{"\"string\"", false, false, &Datapoint{Data: 4}, &Datapoint{Data: "string"}},
		{"\"❤ ☀ ☆ ☂ ☻ ♞ ☯ ☭ ☢ €\"", false, false, &Datapoint{Data: 4}, &Datapoint{Data: "❤ ☀ ☆ ☂ ☻ ♞ ☯ ☭ ☢ €"}},

		// Literal identity
		{"$", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},

		// Basic Testing
		{"4 < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},

		// Logical tests
		{"true or false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"false or false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"true and false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"true and (false or true)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"true and true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"true and not false", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},

		// Logical filter tests
		{"if true", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if true | 42", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 42}},
		{"if false", false, false, &Datapoint{Data: 4}, nil},
		{"if $ < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},

		// Comparison
		{"$ > 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"$ > 3", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ >= 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ < 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"$ < 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ <= 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ != 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"$ != 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ == 4", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"$ == 5", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},

		// Logical pipelines
		{"if $ < 5 and $ > 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if $ < 5 | if $ > 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"if $ < 5 | if $ > 33", false, false, &Datapoint{Data: 4}, nil},
		{"if $ < 5 | $ > 33", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"has(\"test\") | $ < 1", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"if has(\"test\")| $ < 1", false, false, &Datapoint{Data: 4}, nil},
		{"if has(\"test\")| $[\"test\"] < 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: false}},
		{"if has(\"tst\")| $[\"test\"] < 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, nil},
		{"if has(\"test\")| $[\"test\"] > 1", false, false, &Datapoint{Data: map[string]interface{}{"test": 25}}, &Datapoint{Data: true}},

		// Invalid
		{"if has(\"test\"", true, false, nil, nil},
		{"$[\"test\"]", false, true, &Datapoint{Data: 4}, nil},

		// Multiple stage pipeline
		{"$ | false | 42", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 42}},

		// implicit logicals
		{"gt(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"gt(3)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"gte(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"lt(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"lt(5)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"lte(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"ne(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"ne(5)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"eq(4)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"eq(5)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},

		// Test custom functions
		{"identity()", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 4}},
		{"passthrough($ > 5)", false, false, &Datapoint{Data: 4}, &Datapoint{Data: false}},
		{"passthrough($ > 5 | eq(false))", false, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"passthrough($ > 5 | eq(false), $)", true, false, &Datapoint{Data: 4}, &Datapoint{Data: true}},
		{"fortyTwo()", false, false, &Datapoint{Data: 4}, &Datapoint{Data: 42}},
		{"doesnotexist()", true, false, &Datapoint{Data: 4}, nil},
	}

	// function that should nilt out
	identityFunc := func(name string, children ...TransformFunc) (TransformFunc, error) {
		return func(dp *Datapoint) (tdp *Datapoint, err error) {
			return dp, nil
		}, nil
	}
	RegisterCustomFunction("identity", identityFunc)

	// passthrough
	passthroughFunc := func(name string, children ...TransformFunc) (TransformFunc, error) {
		if len(children) != 1 {
			return pipelineGeneratorIdentity(), errors.New("passthrough error")
		}
		return func(dp *Datapoint) (tdp *Datapoint, err error) {
			return children[0](dp)
		}, nil
	}
	RegisterCustomFunction("passthrough", passthroughFunc)

	fortyTwo := func(name string, children ...TransformFunc) (TransformFunc, error) {
		return func(dp *Datapoint) (tdp *Datapoint, err error) {
			dp.Data = 42
			return dp, nil
		}, nil
	}
	RegisterCustomFunction("fortyTwo", fortyTwo)

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
