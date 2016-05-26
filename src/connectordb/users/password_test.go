/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package users

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalcHash(t *testing.T) {
	_, err := calcHash("password", "", "")
	require.Error(t, err)
	h2, err := calcHash("password", "", "SHA512")
	require.NoError(t, err)
	h3, err := calcHash("password", "a", "SHA512")
	require.NoError(t, err)
	h4, err := calcHash("password2", "a", "SHA512")
	require.NoError(t, err)
	h5, err := calcHash("password2", "a", "SHA512")
	require.NoError(t, err)

	if h2 == h3 {
		t.Errorf("h2 and h3 should not match")
	}

	if h3 == h4 {
		t.Errorf("h4 and h3 should not match")
	}

	if h5 != h4 {
		t.Errorf("h5 and h4 should match")
	}

	require.NoError(t, CheckPassword("password", h3, "a", "SHA512"))
	require.Error(t, CheckPassword("password2", h3, "a", "SHA512"))

	h1, err := calcHash("pass", "lol", "bcrypt")
	require.NoError(t, err)

	require.Error(t, CheckPassword("mylol", h1, "lol", "bcrypt"))
	require.NoError(t, CheckPassword("pass", h1, "lol", "bcrypt"))
}
