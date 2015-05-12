package streamdb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {

	tc, err := NewTimedCache(2, 10, nil)
	require.NoError(t, err)

	_, _, ok := tc.GetByID(1)
	require.False(t, ok)

	_, ok = tc.GetByName("hi")
	require.False(t, ok)

	tc.Set("hi", 0, "ho")

	v, name, ok := tc.GetByID(0)
	require.True(t, ok)
	require.Equal(t, "hi", name)
	require.Equal(t, "ho", v.(string))

	v, ok = tc.GetByName("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))

	tc.Set("po", 1, "tato")
	v, ok = tc.GetByName("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))

	tc.Set("yo", 2, "lo")

	//Now potato should be evicted
	v, ok = tc.GetByName("po")
	require.False(t, ok)

	v, ok = tc.GetByName("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))
	v, ok = tc.GetByName("yo")
	require.True(t, ok)
	require.Equal(t, "lo", v.(string))

	tc.UnlinkNamePrefix("y")
	_, ok = tc.GetByName("yo")
	require.False(t, ok)
	v, ok = tc.GetByName("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))

	//Lastly let's check the lime-based eviction by setting the eviction time to 0
	tc, err = NewTimedCache(2, 0, nil)
	tc.Set("hi", 152, "ho")
	v, ok = tc.GetByName("hi")
	require.False(t, ok)

	//check error
	_, err = NewTimedCache(2, 0, errors.New("LOL FALE"))
	require.Error(t, err)

	//Now check all the other cache functinos that are new with the new cache version
	tc, err = NewTimedCache(2, 10, nil)
	require.NoError(t, err)

	tc.SetID(3, "hi")
	val, k, ok := tc.GetByID(3)
	require.True(t, ok)
	require.Equal(t, "", k)
	require.Equal(t, "hi", val.(string))

	tc.Set("ho", 3, "hello")
	val, k, ok = tc.GetByID(3)
	require.True(t, ok)
	require.Equal(t, "ho", k)
	require.Equal(t, "hello", val.(string))

	tc.Update(3, "zoo")
	v, ok = tc.GetByName("ho")
	require.True(t, ok)
	require.Equal(t, "zoo", v.(string))

	tc.RemoveName("ho")
	_, _, ok = tc.GetByID(3)
	require.False(t, ok)

	tc.Set("hp", 3, "hello")
	_, val, ok = tc.GetByID(3)
	require.True(t, ok)
	require.Equal(t, val, "hp")
	tc.UnlinkName("hp")
	_, ok = tc.GetByName("hp")
	require.False(t, ok)
	_, val, ok = tc.GetByID(3)
	require.True(t, ok)
	require.Equal(t, val, "")

	tc.Set("up", 8, "hello")
	tc.Update(8, "hi")
	v, ok = tc.GetByName("up")
	require.True(t, ok)
	require.Equal(t, "hi", v.(string))
}
