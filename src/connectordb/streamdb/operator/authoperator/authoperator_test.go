package authoperator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthOperatorBasics(t *testing.T) {

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	require.NoError(t, baseOperator.CreateUser("streamdb_test", "root@localhost", "mypass"))

	_, err = NewUserLoginOperator(&baseOperator, "streamdb_test", "wrongpass")
	require.Error(t, err)

	o, err := NewUserLoginOperator(&baseOperator, "streamdb_test", "mypass")
	require.NoError(t, err)

	require.Equal(t, "streamdb_test/user", o.Name())

	u, err := o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	d, err := o.Device()
	require.NoError(t, err)
	require.Equal(t, "user", d.Name)

}
