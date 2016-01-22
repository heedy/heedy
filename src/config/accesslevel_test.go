package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccessLevelMap(t *testing.T) {
	a := AccessLevel{UserName: true}
	m := a.GetMap()
	require.False(t, m["user_password"])
	require.True(t, m["user_name"])
}
