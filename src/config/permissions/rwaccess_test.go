package permissions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRWAccessMap(t *testing.T) {
	a := RWAccess{UserName: true}
	m := a.GetMap()
	require.False(t, m["user_password"])
	require.True(t, m["user_name"])
}
