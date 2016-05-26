/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	cfg := NewConfiguration()
	require.NoError(t, cfg.Validate())

	cfg.Permissions = "boo"
	require.Error(t, cfg.Validate())
}

func TestSave(t *testing.T) {
	cfg := NewConfiguration()

	require.NoError(t, cfg.Save("test.conf"))

	cfg2, err := Load("test.conf")
	require.NoError(t, err)
	require.NoError(t, cfg2.Validate())
}
