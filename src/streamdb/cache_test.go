package streamdb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	tc, err := NewTimedCache(2, 10, nil)
	require.NoError(t, err)

	tc.Add("hi", "ho")

	v, ok := tc.Get("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))

	tc.Add("po", "tato")
	v, ok = tc.Get("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))

	tc.Add("yo", "lo")

	//Now potato should be evicted
	v, ok = tc.Get("po")
	require.False(t, ok)

	v, ok = tc.Get("hi")
	require.True(t, ok)
	require.Equal(t, "ho", v.(string))
	v, ok = tc.Get("yo")
	require.True(t, ok)
	require.Equal(t, "lo", v.(string))

	//Lastly let's check the lime-based eviction by setting the eviction time to 0
	tc, err = NewTimedCache(2, 0, nil)
	tc.Add("hi", "ho")
	v, ok = tc.Get("hi")
	require.False(t, ok)

	//Lastly, check error
	_, err = NewTimedCache(2, 0, errors.New("LOL FALE"))
	require.Error(t, err)
}
