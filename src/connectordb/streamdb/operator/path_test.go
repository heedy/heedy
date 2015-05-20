package operator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitStreamPath(t *testing.T) {
	username, devicepath, streampath, streamname, substreams, err := SplitStreamPath("user/device/stream/sub1/sub2", nil)

	require.NoError(t, err)
	require.Equal(t, username, "user")
	require.Equal(t, devicepath, "user/device")
	require.Equal(t, streampath, "user/device/stream")
	require.Equal(t, streamname, "stream")
	require.Equal(t, substreams, "sub1/sub2")

	_, _, _, _, _, err = SplitStreamPath("user/device", nil)
	require.Error(t, err)

	_, _, _, _, _, err = SplitStreamPath("user/device/stream/sub1/sub2", errors.New("FAILE"))
	require.Error(t, err)

	username, devicepath, streampath, streamname, substreams, err = SplitStreamPath("user/device/stream", nil)

	require.NoError(t, err)
	require.Equal(t, username, "user")
	require.Equal(t, devicepath, "user/device")
	require.Equal(t, streampath, "user/device/stream")
	require.Equal(t, streamname, "stream")
	require.Equal(t, substreams, "")
}

func TestSplitDevicePath(t *testing.T) {
	u, d, err := SplitDevicePath("user/device", nil)

	require.NoError(t, err)
	require.Equal(t, u, "user")
	require.Equal(t, d, "device")

	_, _, err = SplitDevicePath("user/device/something", nil)
	require.Error(t, err)

	_, _, err = SplitDevicePath("user/device", errors.New("FAIL"))
	require.Error(t, err)
}
