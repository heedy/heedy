/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package users

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountUsers(t *testing.T) {

	for _, testdb := range testdatabases {
		before, err := testdb.CountUsers()
		require.Nil(t, err)

		_, _, _, err = CreateUDS(testdb)
		require.Nil(t, err)

		after, err := testdb.CountUsers()
		require.Nil(t, err)

		fmt.Printf("user before: %v after: %v\n", before, after)
		require.Equal(t, before+1, after)
	}
}

func TestCountStreams(t *testing.T) {

	for _, testdb := range testdatabases {
		before, err := testdb.CountStreams()
		require.Nil(t, err)

		_, _, _, err = CreateUDS(testdb)
		require.Nil(t, err)

		after, err := testdb.CountStreams()
		require.Nil(t, err)
		fmt.Printf("streams before: %v after: %v\n", before, after)

		require.True(t, before < after)
	}
}

func TestCountDevices(t *testing.T) {

	for _, testdb := range testdatabases {
		before, err := testdb.CountDevices()
		require.Nil(t, err)

		_, _, _, err = CreateUDS(testdb)
		require.Nil(t, err)

		after, err := testdb.CountDevices()
		require.Nil(t, err)
		fmt.Printf("dev before: %v after: %v\n", before, after)

		require.True(t, before < after)
	}
}
