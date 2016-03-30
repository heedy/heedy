package users

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateStruct(t *testing.T) {
	for _, testdb := range testdatabases {
		require.NoError(t, testdb.CreateUser(&UserMaker{
			User: User{Name: "structtest", Email: "test@test.com", Password: "mypass", Role: "user"},
			Streams: map[string]*StreamMaker{
				"stream1": &StreamMaker{Stream: Stream{
					Schema: `{"type":"string"}`,
				}},
			},
			Devices: map[string]*DeviceMaker{
				"dev1": &DeviceMaker{Streams: map[string]*StreamMaker{
					"devstream": &StreamMaker{Stream: Stream{
						Schema: `{"type": "number"}`,
					}},
				}},
			},
		}))

		// Make sure the user exists
		u, err := testdb.ReadUserByName("structtest")
		require.NoError(t, err)

		d, err := testdb.ReadDeviceForUserByName(u.UserID, "dev1")
		require.NoError(t, err)

		s, err := testdb.ReadStreamByDeviceIDAndName(d.DeviceID, "devstream")
		require.NoError(t, err)

		require.Equal(t, s.Schema, `{"type": "number"}`)

		d, err = testdb.ReadDeviceForUserByName(u.UserID, "user")
		require.NoError(t, err)

		s, err = testdb.ReadStreamByDeviceIDAndName(d.DeviceID, "stream1")
		require.NoError(t, err)

		require.Equal(t, s.Schema, `{"type":"string"}`)

	}
}
