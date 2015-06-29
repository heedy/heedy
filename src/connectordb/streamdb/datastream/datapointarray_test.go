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
	dpa7 = DatapointArray{
		Datapoint{1., "test0", ""},
		Datapoint{2., "test1", ""},
		Datapoint{3., "test2", ""},
		Datapoint{4., "test3", ""},
		Datapoint{5., "test4", ""},
		Datapoint{6., "test5", ""},
		Datapoint{6., "test6", ""},
		Datapoint{7., "test7", ""},
		Datapoint{8., "test8", ""},
	}

	dpa8 = DatapointArray{Datapoint{2.0, "helloWorld", "me"}, Datapoint{1.0, "helloWorld2", "me2"}}
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

	dpat, err := DatapointArrayFromBytes(dpb)
	require.NoError(t, err)

	require.True(t, dpa1.IsEqual(dpat))

	dpb, err = dpa4.Bytes()
	require.NoError(t, err)

	dpat, err = DatapointArrayFromBytes(dpb)
	require.NoError(t, err)

	require.True(t, dpa4.IsEqual(dpat))

	dpb, err = dpa5.Bytes()
	require.NoError(t, err)

	dpat, err = DatapointArrayFromBytes(dpb)
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

	d1, err := DatapointArrayFromBytes(dbytea[0])
	require.NoError(t, err)
	d2, err := DatapointArrayFromBytes(dbytea[1])
	require.NoError(t, err)
	d3, err := DatapointArrayFromBytes(dbytea[2])
	require.NoError(t, err)

	require.True(t, d1.IsEqual(dpa6[0:2]))
	require.True(t, d2.IsEqual(dpa6[2:4]))
	require.True(t, d3.IsEqual(dpa6[4:5]))

	dbytea, err = dpa1.SplitIntoChunks(2)
	require.NoError(t, err)
	require.Equal(t, 1, len(dbytea))
	d1, err = DatapointArrayFromBytes(dbytea[0])
	require.True(t, d1.IsEqual(dpa1[0:2]))
}

func TestDatapointArrayEncodeDecode(t *testing.T) {
	_, err := dpa1.Encode(3)
	require.Error(t, err)

	da, err := dpa1.Encode(MsgPackVersion)
	require.NoError(t, err)

	dpa, err := DecodeDatapointArray(da, 3)
	require.Error(t, err)

	dpa, err = DecodeDatapointArray(da, 2)
	require.Error(t, err)

	dpa, err = DecodeDatapointArray(da, MsgPackVersion)
	require.NoError(t, err)
	require.Equal(t, dpa.String(), dpa1.String())

	da, err = dpa1.Encode(CompressedMsgPackVersion)
	require.NoError(t, err)

	dpa, err = DecodeDatapointArray(da, CompressedMsgPackVersion)
	require.NoError(t, err)
	require.Equal(t, dpa.String(), dpa1.String())
}

func TestTimestampOrdered(t *testing.T) {
	require.True(t, dpa7.IsTimestampOrdered())
	require.False(t, dpa8.IsTimestampOrdered())
}

func TestTimeIndex(t *testing.T) {
	require.Equal(t, DatapointArray{}.FindTimeIndex(1.0), -1)
	require.Equal(t, -1, dpa7.FindTimeIndex(20.0))
	require.Equal(t, 3, dpa7.FindTimeIndex(3.0))
	require.Equal(t, 7, dpa7.FindTimeIndex(6.0))
}

func TestTRange(t *testing.T) {
	require.Equal(t, dpa7.TRange(0.5, 6.0).String(), dpa7[:7].String())
	require.Equal(t, dpa7.TRange(0.5, 80.0).String(), dpa7.String())
}
