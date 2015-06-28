package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

var (
	dpa1 = DatapointArray{Datapoint{1.0, "helloWorld", "me"}, Datapoint{2.0, "helloWorld2", "me2"}}
	dpa2 = DatapointArray{Datapoint{1.0, "helloWorl", "me"}, Datapoint{2.0, "helloWorld2", "me2"}}
	dpa3 = DatapointArray{Datapoint{1.0, "helloWorl", "me"}}

	dpa4 = DatapointArray{Datapoint{3.0, 12.0, ""}}

	//Warning: the map types change depending on marshaller/unmarshaller is used
	dpa5 = DatapointArray{Datapoint{3.0, map[interface{}]interface{}{"hello": 2.0, "y": "hi"}, ""}}

	dpa6 = DatapointArray{Datapoint{1.0, 1.0, ""}, Datapoint{2.0, 2.0, ""}, Datapoint{3.0, 3., ""}, Datapoint{4.0, 4., ""}, Datapoint{5.0, 5., ""}}
)

func TestDatapointArrayString(t *testing.T) {
	require.Equal(t, "DatapointArray{[T=1.000 D=helloWorl S=me]}", dpa3.String())
}

func TestEquality(t *testing.T) {
	require.True(t, dpa1.IsEqual(dpa1))
	require.False(t, dpa1.IsEqual(dpa2))
	require.False(t, dpa2.IsEqual(dpa3))
	require.True(t, dpa4.IsEqual(dpa4))
	require.True(t, dpa5.IsEqual(dpa5))
}

func TestDatapointArrayBytes(t *testing.T) {
	dpb, err := dpa1.Bytes()
	require.NoError(t, err)

	dpat, err := LoadDatapointArray(dpb)
	require.NoError(t, err)

	require.True(t, dpa1.IsEqual(dpat))

	dpb, err = dpa4.Bytes()
	require.NoError(t, err)

	dpat, err = LoadDatapointArray(dpb)
	require.NoError(t, err)

	require.True(t, dpa4.IsEqual(dpat))

	dpb, err = dpa5.Bytes()
	require.NoError(t, err)

	dpat, err = LoadDatapointArray(dpb)
	require.NoError(t, err)

	require.True(t, dpa5.IsEqual(dpat))

}

func TestDatapointArraySchema(t *testing.T) {
	sString, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{"type":"string"}`))
	require.NoError(t, err)
	sFloat, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{"type":"number"}`))
	require.NoError(t, err)
	sObj, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(`{"type": "object", "properties": {"lat": {"type": "number"},"msg": {"type": "string"}}}`))

	vString := "Hello"
	vFloat := 3.14
	vObj := map[string]interface{}{"lat": 88.32, "msg": "hi"}
	vObjBad := map[string]interface{}{"lat": "88.32", "msg": "hi"}
	//vObjBadExtra := map[string]interface{}{"lat": 88.32, "msg": "hi", "testing": 123}

	testDpa := DatapointArray{Datapoint{1.0, vString, ""}, Datapoint{1.0, vFloat, ""}}

	require.Error(t, testDpa.VerifySchema(sString))

	testDpa[1].Data = "hai!"
	require.NoError(t, testDpa.VerifySchema(sString))

	require.Error(t, testDpa.VerifySchema(sFloat))
	testDpa[1].Data = vFloat
	testDpa[0].Data = "10.0"
	require.Error(t, testDpa.VerifySchema(sFloat))
	testDpa[0].Data = 10.0
	require.NoError(t, testDpa.VerifySchema(sFloat))

	testDpa[0].Data = vObj
	testDpa[1].Data = vObj
	require.NoError(t, testDpa.VerifySchema(sObj))
	testDpa[1].Data = vObjBad
	require.Error(t, testDpa.VerifySchema(sObj))
	//testDpa[1].Data = vObjBadExtra					//TODO: GoJsonSchema does not fail on exta fields!
	//require.Error(t, testDpa.VerifySchema(sObj))

}

func TestDatapointArrayChunks(t *testing.T) {
	dbytea, err := dpa6.SplitIntoChunks(2)
	require.NoError(t, err)
	require.Equal(t, 3, len(dbytea))

	d1, err := LoadDatapointArray(dbytea[0])
	require.NoError(t, err)
	d2, err := LoadDatapointArray(dbytea[1])
	require.NoError(t, err)
	d3, err := LoadDatapointArray(dbytea[2])
	require.NoError(t, err)

	require.True(t, d1.IsEqual(dpa6[0:2]))
	require.True(t, d2.IsEqual(dpa6[2:4]))
	require.True(t, d3.IsEqual(dpa6[4:5]))

	dbytea, err = dpa1.SplitIntoChunks(2)
	require.NoError(t, err)
	require.Equal(t, 1, len(dbytea))
	d1, err = LoadDatapointArray(dbytea[0])
	require.True(t, d1.IsEqual(dpa1[0:2]))
}
