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
	// We don't want to modify the original default structure, since other tests require it.
	require.NoError(t, Default.Save("permissions.conf"))

	cfg, err := Load("permissions.conf")
	require.NoError(t, err)

	require.NoError(t, cfg.Validate())

	p := cfg.UserRoles["user"]
	p.PublicAccessLevel = "lol"
	cfg.UserRoles["user"] = p
	require.Error(t, cfg.Validate())

	p.PublicAccessLevel = "none"
	cfg.UserRoles["user"] = p
	require.NoError(t, cfg.Validate())

	delete(cfg.UserRoles, "user")
	require.Error(t, cfg.Validate())

	cfg.UserRoles["user"] = p
	require.NoError(t, cfg.Validate())

	delete(cfg.UserRoles, "nobody")
	require.Error(t, cfg.Validate())
}
