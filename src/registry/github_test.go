package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithub(t *testing.T) {
	g := NewGithubClient("")

	_, err := g.Get("gerhub.com/dkumor/test")
	require.Error(t, err)

	_, err = g.Get("http://gerhub.com/dkumor/test")
	require.Error(t, err)

	_, err = g.Get("http://github.com/")
	require.Error(t, err)

	p, err := g.Get("http://github.com/connectordb/cdb-analysis")
	// require.Error(t, err, "ConnectorDB's default config is in asset folder, but does not define any plugins")

	require.NoError(t, err)
	require.Equal(t, p.Stars, 157)

}
