/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package rediscache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisCache(t *testing.T) {
	require.NoError(t, rc.Clear())

	rc.BatchSize = 2

	r := RedisCache{rc}

	i, err := r.StreamLength(1, 2, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 0, i)

	i, err = r.Insert(1, 2, "hi", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	i, err = r.StreamLength(1, 2, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	b, err := r.ReadProcessingQueue()
	require.NoError(t, err)
	require.Nil(t, b)

	b, err = r.ReadBatches(1)
	require.NoError(t, err)

	b2, err := r.ReadProcessingQueue()
	require.NoError(t, err)
	require.EqualValues(t, b, b2)

	require.EqualValues(t, 1, len(b))
	id, err := b[0].GetDeviceID()
	require.NoError(t, err)
	require.EqualValues(t, 1, id)
	id, _ = b[0].GetStreamID()
	require.EqualValues(t, 2, id)
	require.EqualValues(t, "hi", b[0].Substream)
	require.EqualValues(t, dpa6[:2].String(), b[0].Data.String())

	require.NoError(t, r.ClearBatches(b))

	b, err = r.ReadProcessingQueue()
	require.NoError(t, err)
	require.Nil(t, b)

	i, err = r.StreamLength(1, 2, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	dpa, _, _, err := r.ReadRange(1, 2, "hi", 1, 2)
	require.NoError(t, err)
	require.Nil(t, dpa)

	dpa, _, _, err = r.ReadRange(1, 2, "hi", 2, 3)
	require.NoError(t, err)
	require.EqualValues(t, dpa6[2:3].String(), dpa.String())

	rc.BatchSize = 250
}

func TestRedisCacheDelete(t *testing.T) {
	require.NoError(t, rc.Clear())

	rc.BatchSize = 2

	r := RedisCache{rc}
	i, err := r.Insert(1, 2, "hi", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)
	i, err = r.Insert(1, 2, "ho", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)
	i, err = r.Insert(1, 3, "hi", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)
	i, err = r.Insert(2, 3, "hi", dpa6, false)
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	require.NoError(t, r.DeleteSubstream(1, 2, "hi"))
	i, err = r.StreamLength(1, 2, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 0, i)
	i, err = r.StreamLength(1, 2, "ho")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)
	i, err = r.StreamLength(1, 3, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	require.NoError(t, r.DeleteStream(1, 2))
	i, err = r.StreamLength(1, 2, "ho")
	require.NoError(t, err)
	require.EqualValues(t, 0, i)
	i, err = r.StreamLength(1, 3, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)

	require.NoError(t, r.DeleteDevice(1))
	i, err = r.StreamLength(1, 3, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 0, i)
	i, err = r.StreamLength(2, 3, "hi")
	require.NoError(t, err)
	require.EqualValues(t, 5, i)
}
