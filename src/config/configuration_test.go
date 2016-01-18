/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	cfg := NewConfiguration()

	require.True(t, cfg.IsAllowedEmail("foo@bar.com"))
	cfg.AllowedEmailSuffixes = []string{"bar.com", "baz.com"}

	require.False(t, cfg.IsAllowedEmail("foo@foo.com"))
	require.True(t, cfg.IsAllowedEmail("foo@bar.com"))
	require.True(t, cfg.IsAllowedEmail("foo@baz.com"))

	require.True(t, cfg.IsAllowedEmail("foo@subdomain.baz.com"))
}

func TestValidate(t *testing.T) {
	cfg := NewConfiguration()
	require.NoError(t, cfg.Validate())

	p := cfg.Permissions["user"]
	p.PublicReadAccessLevel = "lol"
	cfg.Permissions["user"] = p
	require.Error(t, cfg.Validate())

	delete(cfg.Permissions, "user")
	require.Error(t, cfg.Validate())
}

func TestSave(t *testing.T) {
	cfg := NewConfiguration()

	require.NoError(t, cfg.Save("test.conf"))

	cfg2, err := Load("test.conf")
	require.NoError(t, err)
	require.NoError(t, cfg2.Validate())
}
