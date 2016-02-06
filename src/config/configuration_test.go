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
