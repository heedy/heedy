package streamdb

import (
	"connectordb/streamdb/operator"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthSubscribe(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()

	//Let's create a stream
	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/tst"))
	require.NoError(t, db.CreateDevice("tst/tst2"))
	require.NoError(t, db.CreateStream("tst/tst/tst", `{"type": "string"}`))

	recvchan := make(chan operator.Message, 2)
	recvchan2 := make(chan operator.Message, 2)
	recvchan3 := make(chan operator.Message, 2)

	o, err := db.GetOperator("tst/tst2")
	require.NoError(t, err)

	_, err = o.Subscribe("tst", recvchan)
	require.Error(t, err)
	_, err = o.Subscribe("tst/tst", recvchan)
	require.Error(t, err)
	_, err = o.Subscribe("tst/tst/tst", recvchan)
	require.Error(t, err)

	o, err = db.GetOperator("tst/tst")
	require.NoError(t, err)

	_, err = o.Subscribe("tst", recvchan)
	require.Error(t, err)

	_, err = o.Subscribe("tst/tst", recvchan2)
	require.NoError(t, err)
	_, err = o.Subscribe("tst/tst/tst", recvchan3)
	require.NoError(t, err)

	db.SetAdmin("tst/tst", true) //TODO: Subscriptions should be dumped on a permissions change, and that does not happen
	_, err = o.Subscribe("tst", recvchan)
	require.NoError(t, err)

	db.msg.Flush()

	data := []operator.Datapoint{operator.Datapoint{
		Timestamp: 1.0,
		Data:      "Hello World!",
	}}
	require.NoError(t, o.InsertStream("tst/tst/tst", data))
	//We bind a timeout to the channel, since we want the test to fail if no messages come through
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- operator.Message{"TIMEOUT", []operator.Datapoint{}}
		recvchan2 <- operator.Message{"TIMEOUT", []operator.Datapoint{}}
		recvchan3 <- operator.Message{"TIMEOUT", []operator.Datapoint{}}
	}()

	m := <-recvchan
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")
	m = <-recvchan2
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")
	m = <-recvchan3
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")
}
