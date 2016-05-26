/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitStreamPath(t *testing.T) {
	username, devicepath, streampath, streamname, substreams, err := SplitStreamPath("user/device/stream/sub1/sub2")

	require.NoError(t, err)
	require.Equal(t, username, "user")
	require.Equal(t, devicepath, "user/device")
	require.Equal(t, streampath, "user/device/stream")
	require.Equal(t, streamname, "stream")
	require.Equal(t, substreams, "sub1/sub2")

	_, _, _, _, _, err = SplitStreamPath("user/device")
	require.Error(t, err)

	username, devicepath, streampath, streamname, substreams, err = SplitStreamPath("user/device/stream")

	require.NoError(t, err)
	require.Equal(t, username, "user")
	require.Equal(t, devicepath, "user/device")
	require.Equal(t, streampath, "user/device/stream")
	require.Equal(t, streamname, "stream")
	require.Equal(t, substreams, "")
}

func TestSplitDevicePath(t *testing.T) {
	u, d, err := SplitDevicePath("user/device")

	require.NoError(t, err)
	require.Equal(t, u, "user")
	require.Equal(t, d, "device")

	_, _, err = SplitDevicePath("user/device/something")
	require.Error(t, err)

}

func ExamplePath_IsUser() {
	path, _ := CreatePath("username")
	fmt.Printf("%v\n", path.IsUser())

	path, _ = CreatePath("username/devicename")
	fmt.Printf("%v\n", path.IsUser())

	// Output:
	// true
	// false
}

func ExamplePath_IsDevice() {
	path, _ := CreatePath("username")
	fmt.Printf("%v\n", path.IsDevice())

	path, _ = CreatePath("username/devicename")
	fmt.Printf("%v\n", path.IsDevice())

	path, _ = CreatePath("username/devicename/")
	fmt.Printf("%v\n", path.IsDevice())

	// Output:
	// false
	// true
	// false
}

func ExamplePath_IsStream() {
	path, _ := CreatePath("user/dev")
	fmt.Printf("%v\n", path.IsStream())

	path, _ = CreatePath("user/device/stream")
	fmt.Printf("%v\n", path.IsStream())

	path, _ = CreatePath("user/device/stream/substream")
	fmt.Printf("%v\n", path.IsStream())

	// Output:
	// false
	// true
	// false
}

func ExamplePath_IsSubstream() {
	path, _ := CreatePath("user/dev/stream")
	fmt.Printf("%v\n", path.IsSubstream())

	path, _ = CreatePath("user/device/stream/substream")
	fmt.Printf("%v\n", path.IsSubstream())

	// Output:
	// false
	// true
}

func ExamplePath_GetSubstreamPath() {
	path, _ := CreatePath("user/device/stream/substream")
	fmt.Printf("%v\n", path.GetSubstreamPath())
	// Output: user/device/stream/substream

}
