package authoperator

/**
func TestAuthStreamIO(t *testing.T) {

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	//Let's create a stream
	require.NoError(t, baseOperator.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, baseOperator.CreateDevice("tst/tst"))

	ao, err := NewDeviceAuthOperator(baseOperator, "tst/tst")
	require.NoError(t, err)
	o := interfaces.PathOperatorMixin{ao}

	require.NoError(t, o.CreateStream("tst/tst/tst", `{"type": "integer"}`))

	//Now make sure that length is 0
	l, err := o.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	strm, err := o.ReadStream("tst/tst/tst")
	require.NoError(t, err)
	l, err = o.LengthStreamByID(strm.StreamId, "")

	data := []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.NoError(t, o.InsertStream("tst/tst/tst", data, false))

	l, err = o.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(1), l)

	dr, err := o.GetStreamTimeRange("tst/tst/tst", 0.0, 2.5, 0)
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	dr, err = o.GetStreamIndexRange("tst/tst/tst", 0, 1)
	require.NoError(t, err)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	i, err := baseOperator.TimeToIndexStream("tst/tst/tst", 0.3)
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	//Now let's make sure that stuff is deleted correctly
	require.NoError(t, o.DeleteStream("tst/tst/tst"))
	require.NoError(t, baseOperator.CreateStream("tst/tst/tst", `{"type": "string"}`))
	l, err = baseOperator.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l, "Timebatch has residual data from deleted stream")
}

func TestAuthSubstream(t *testing.T) {
	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	//Let's create a stream
	require.NoError(t, baseOperator.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, baseOperator.CreateDevice("tst/tst"))
	require.NoError(t, baseOperator.CreateDevice("tst/tst2"))
	require.NoError(t, baseOperator.CreateStream("tst/tst2/tst", `{"type": "integer"}`))
	s, err := baseOperator.ReadStream("tst/tst2/tst")
	require.NoError(t, err)
	s.Downlink = true
	require.NoError(t, baseOperator.UpdateStream(s))

	require.NoError(t, baseOperator.SetAdmin("tst/tst", true))

	ao, err := NewDeviceAuthOperator(baseOperator, "tst/tst")
	require.NoError(t, err)
	o := interfaces.PathOperatorMixin{ao}

	data := []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.NoError(t, o.InsertStream("tst/tst2/tst", data, false))

	l, err := o.LengthStream("tst/tst2/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	dr, err := o.GetStreamTimeRange("tst/tst2/tst/downlink", 0.0, 2.5, 0)
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "tst/tst", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

}
**/
