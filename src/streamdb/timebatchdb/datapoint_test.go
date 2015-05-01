package timebatchdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatapoint(t *testing.T) {
	d := NewDatapoint(1337, []byte("Hello World!"), "world/hello")

	if d.Len() != len(d.Bytes()) || d.Timestamp() != 1337 || string(d.Data()) != "Hello World!" || d.DataLen() != 12 || d.Key() != "world/hello" {
		t.Errorf("Datapoint read error: %s", d)
		return
	}

	//Now check that the bytes can be reread from a file
	buf := new(bytes.Buffer)
	buf.Write(d.Bytes())
	buf.Write([]byte("END"))

	d2, err := ReadDatapoint(buf)
	require.NoError(t, err)
	require.Equal(t, d2.Len(), d.Len())
	require.Equal(t, int64(1337), d2.Timestamp())
	require.Equal(t, "Hello World!", string(d2.Data()))
	require.Equal(t, 12, d2.DataLen())
	require.Equal(t, "world/hello", d2.Key())

	require.Equal(t, "END", string(buf.Next(3)), "Datapoint reading went out of bounds")

	//Lastly, check if the datapoint can be created from a byte array
	buf = new(bytes.Buffer)
	buf.Write(d.Bytes())
	buf.Write([]byte("END"))
	d3, n := DatapointFromBytes(buf.Bytes())
	require.Equal(t, d3.Len(), int(n))

	require.Equal(t, d3.Len(), d.Len())
	require.Equal(t, int64(1337), d3.Timestamp())
	require.Equal(t, "Hello World!", string(d3.Data()))
	require.Equal(t, 12, d3.DataLen())
	require.Equal(t, "world/hello", d3.Key())
	require.Equal(t, d.String(), d3.String())
}

func TestLargeDatapoint(t *testing.T) {
	datastring := "Hello World"

	for i := 0; i < 10000; i++ {
		datastring = datastring + "HelloWorld"
	}
	d := NewDatapoint(1337, []byte(datastring), "")
	require.Equal(t, len(d.Bytes()), d.Len())
	require.Equal(t, int64(1337), d.Timestamp())
	require.Equal(t, string(d.Data()), datastring)
	require.Equal(t, d.DataLen(), len(datastring))
	require.Equal(t, "", d.Key())
}

func BenchmarkDatapointCreate(b *testing.B) {
	//Create the datapoint and get the Bytes from it
	for n := 0; n < b.N; n++ {
		d := NewDatapoint(1337, []byte("Hello World!"), "")
		d.Bytes()
	}
}

func BenchmarkDatapointRead(b *testing.B) {
	//Read stuff from the datapoint
	d := NewDatapoint(1337, []byte("Hello World!"), "")
	for n := 0; n < b.N; n++ {
		d.Timestamp()
		d.Data()
		d.Key()
	}
}
