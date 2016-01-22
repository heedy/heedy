package permissions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	cfg := &Default

	require.True(t, cfg.IsAllowedEmail("foo@bar.com"))
	cfg.AllowedEmailSuffixes = []string{"bar.com", "baz.com"}

	require.False(t, cfg.IsAllowedEmail("foo@foo.com"))
	require.True(t, cfg.IsAllowedEmail("foo@bar.com"))
	require.True(t, cfg.IsAllowedEmail("foo@baz.com"))

	require.True(t, cfg.IsAllowedEmail("foo@subdomain.baz.com"))
}

func TestSave(t *testing.T) {

	require.NoError(t, Default.Save("permissions.conf"))

	cfg2, err := Load("permissions.conf")
	require.NoError(t, err)
	require.NoError(t, cfg2.Validate())
}

func TestValidate(t *testing.T) {
	cfg := &Default
	require.NoError(t, cfg.Validate())

	p := cfg.Roles["user"]
	p.PublicReadAccessLevel = "lol"
	cfg.Roles["user"] = p
	require.Error(t, cfg.Validate())

	delete(cfg.Roles, "user")
	require.NoError(t, cfg.Validate())
	delete(cfg.Roles, "nobody")
	require.Error(t, cfg.Validate())
}
