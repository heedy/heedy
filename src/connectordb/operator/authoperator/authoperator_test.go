/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator

import (
	"connectordb/operator/interfaces"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthOperatorBasics(t *testing.T) {
	fmt.Println("test authoperator basics")

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	require.NoError(t, baseOperator.CreateUser("streamdb_test", "root@localhost", "mypass"))

	_, err = NewUserLoginOperator(baseOperator, "streamdb_test", "wrongpass")
	require.Error(t, err)

	o, err := NewUserLoginOperator(baseOperator, "streamdb_test", "mypass")
	require.NoError(t, err)

	require.Equal(t, "streamdb_test/user", o.Name())

	u, err := o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	d, err := o.Device()
	require.NoError(t, err)
	require.Equal(t, "user", d.Name)

	apiOp, err := NewAPILoginOperator(baseOperator, d.ApiKey)
	require.NoError(t, err)
	require.Equal(t, "streamdb_test/user", apiOp.Name())

}

func TestCountAll(t *testing.T) {

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	require.NoError(t, baseOperator.CreateUser("streamdb_test", "root@localhost", "mypass"))

	userLoginOperator, err := NewUserLoginOperator(baseOperator, "streamdb_test", "mypass")
	require.NoError(t, err)
	o := interfaces.PathOperatorMixin{userLoginOperator}

	_, err = o.CountUsers()
	require.Error(t, err)
	_, err = o.CountDevices()
	require.Error(t, err)
	_, err = o.CountStreams()
	require.Error(t, err)

	baseOperator.SetAdmin("streamdb_test", true)

	i, err := o.CountUsers()
	require.NoError(t, err)
	require.EqualValues(t, i, 1)
	i, err = o.CountDevices()
	require.NoError(t, err)
	require.True(t, i >= 1)
	i, err = o.CountStreams()
	require.NoError(t, err)
	require.True(t, i >= 1)
}
