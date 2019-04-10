package streams

import (
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dpa1 = DatapointArray{&Datapoint{1.0, "helloWorld", "me"}, &Datapoint{2.0, "helloWorld2", "me2"}}
	dpa2 = DatapointArray{&Datapoint{1.0, "helloWorl", "me"}, &Datapoint{2.0, "helloWorld2", "me2"}}
	dpa3 = DatapointArray{&Datapoint{1.0, "helloWorl", "me"}}

	dpa4 = DatapointArray{&Datapoint{3.0, 12.0, ""}}

	//Warning: the map types change depending on marshaller/unmarshaller is used
	dpa5 = DatapointArray{&Datapoint{3.0, map[string]interface{}{"hello": 2.0, "y": "hi"}, ""}}

	dpa6 = DatapointArray{&Datapoint{1.0, 1.0, ""}, &Datapoint{2.0, 2.0, ""}, &Datapoint{3.0, 3., ""}, &Datapoint{4.0, 4., ""}, &Datapoint{5.0, 5., ""}}
	dpa7 = DatapointArray{
		&Datapoint{1., "test0", ""},
		&Datapoint{2., "test1", ""},
		&Datapoint{3., "test2", ""},
		&Datapoint{4., "test3", ""},
		&Datapoint{5., "test4", ""},
		&Datapoint{6., "test5", ""},
		&Datapoint{6., "test6", ""},
		&Datapoint{7., "test7", ""},
		&Datapoint{8., "test8", ""},
	}
)

func TestArrayEquality(t *testing.T) {
	require.True(t, dpa1.IsEqual(dpa1))
	require.False(t, dpa1.IsEqual(dpa2))
	require.False(t, dpa2.IsEqual(dpa3))
	require.True(t, dpa4.IsEqual(dpa4))
	require.True(t, dpa5.IsEqual(dpa5))
}
