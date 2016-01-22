package permissions

import (
	"config"
	"connectordb/users"
	"testing"

	"github.com/stretchr/testify/require"
)

var testUser = users.User{
	UserID:      32,
	Name:        "Daniel",
	Nickname:    "Noob",
	Email:       "my@mail.com",
	Description: "I have no idea what I'm doing",
	Icon:        "myicon",
	Permissions: "user",
	Public:      true,
	Password:    "mypass",
}

var testUser2 = users.User{
	UserID:      33,
	Name:        "Joseph",
	Nickname:    "MyNicknae",
	Email:       "my@mail2.com",
	Description: "Hi!",
	Icon:        "myicon2",
	Permissions: "user",
	Public:      true,
	Password:    "mypass2",
}

var testDevice = users.Device{
	Name:        "MyDevice",
	Nickname:    "woo",
	Description: "super awesome device",
	Icon:        "myicon3",
	UserID:      32,
	DeviceID:    66,
	APIKey:      "helloworld",
	Enabled:     true,
	Public:      true,

	CanReadUser:      true,
	CanReadExternal:  true,
	CanWriteUser:     true,
	CanWriteExternal: true,

	IsVisible:    true,
	UserEditable: true,
}

func TestUserRead(t *testing.T) {
	u := &testUser
	d := &testDevice
	muX := testUser2
	mu := &muX

	m := ReadUserToMap(&config.TestConfiguration, u, d, mu)
	require.NotNil(t, m)

	// The default testing configuration has public read disallow password and permissions
	_, ok := m["permissions"]
	require.False(t, ok)
	_, ok = m["password"]
	require.False(t, ok)
	v, ok := m["description"]
	require.True(t, ok)
	require.Equal(t, muX.Description, v.(string))

	// Now read self
	m = ReadUserToMap(&config.TestConfiguration, u, d, u)
	require.NotNil(t, m)

	_, ok = m["password"]
	require.False(t, ok)
	v, ok = m["permissions"]
	require.True(t, ok)
	require.Equal(t, u.Permissions, v.(string))

	// Finally, read private
	mu.Public = false
	m = ReadUserToMap(&config.TestConfiguration, u, d, mu)
	require.Nil(t, m)

	mu.Permissions = "ocrap"
	m = ReadUserToMap(&config.TestConfiguration, mu, d, mu)
	require.Nil(t, m)
}

func TestUserWrite(t *testing.T) {
	u := &testUser
	d := &testDevice
	muX := testUser2
	mu := &muX

	require.Error(t, UpdateUserFromMap(&config.TestConfiguration, u, d, mu, map[string]interface{}{"name": "hi"}))
	require.Error(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"permissions": "hi"}))

	require.NoError(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"password": "hi"}))
	require.NotEqual(t, "mypass2", mu.Password)

	mu.Permissions = "admin"
	require.Error(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"permissions": "blah"}))
	mu.Permissions = "admin"
	require.NoError(t, UpdateUserFromMap(&config.TestConfiguration, mu, d, mu, map[string]interface{}{"permissions": "user"}))
	require.Equal(t, "user", mu.Permissions)
}
