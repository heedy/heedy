/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator

import (
	"connectordb/operator/interfaces"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthStreamCrud(t *testing.T) {
	fmt.Println("test authstream crud")

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	_, err = baseOperator.ReadAllStreams("bad/badder")
	require.Error(t, err)

	require.NoError(t, baseOperator.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, baseOperator.CreateDevice("tst/testdevice"))

	require.NoError(t, baseOperator.CreateDevice("tst/testdevice2"))
	require.NoError(t, baseOperator.CreateStream("tst/testdevice2/teststream", `{"type":"string"}`))

	ao, err := NewDeviceAuthOperator(baseOperator, "tst/testdevice")
	require.NoError(t, err)
	o := interfaces.PathOperatorMixin{ao}

	dev, err := baseOperator.ReadDevice("tst/testdevice2")
	require.NoError(t, err)
	_, err = o.ReadAllStreamsByDeviceID(dev.DeviceId)
	require.Error(t, err)

	_, err = o.ReadAllStreams("tst/testdevice2")
	require.Error(t, err)

	dev, err = o.Device()
	require.NoError(t, err)
	strms, err := o.ReadAllStreamsByDeviceID(dev.DeviceId)
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	strms, err = o.ReadAllStreams("tst/testdevice")
	require.NoError(t, err)
	require.Equal(t, 0, len(strms))

	require.Error(t, o.CreateStream("tst/testdevice2/mystream", `{"type":"string"}`))
	require.NoError(t, o.CreateStream("tst/testdevice/mystream", `{"type":"string"}`))

	_, err = o.ReadStream("tst/testdevice2/teststream")
	require.Error(t, err)

	s, err := o.ReadStream("tst/testdevice/mystream")
	require.NoError(t, err)
	require.Equal(t, "mystream", s.Name)

	s.Name = "stream2"
	require.NoError(t, o.UpdateStream(s))

	s, err = o.ReadStream("tst/testdevice/mystream")
	require.Error(t, err)

	s, err = baseOperator.ReadStream("tst/testdevice/stream2")
	require.NoError(t, err)
	require.Equal(t, "stream2", s.Name)

	require.Error(t, o.DeleteStream("tst/testdevice2/teststream"))
	require.NoError(t, o.DeleteStream("tst/testdevice/stream2"))

	_, err = baseOperator.ReadStream("tst/testdevice/stream2")
	require.Error(t, err)

	dev, err = o.ReadDevice("tst/testdevice")
	require.NoError(t, err)

	require.NoError(t, o.CreateStreamByDeviceID(dev.DeviceId, "testme", `{"type":"string"}`))

	s, err = o.ReadStreamByDeviceID(dev.DeviceId, "testme")
	require.NoError(t, err)
	require.Equal(t, s.Name, "testme")
	require.NoError(t, o.DeleteStreamByID(s.StreamId, ""))
	_, err = o.ReadStreamByID(s.StreamId)
	require.Error(t, err)
	_, err = o.ReadStreamByDeviceID(dev.DeviceId, "testme")
	require.Error(t, err)
}
