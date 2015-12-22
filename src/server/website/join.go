/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"config"
	"connectordb"
	"errors"
	"net/http"
	"server/webcore"
)

func checkIfJoinAllowed(request *http.Request) error {
	joinpermissions := "nobody"
	// First check if the user is authenticated (ie, a user is trying to add another user)
	o, err := webcore.Authenticate(Database, request)
	if err == nil {
		// Auth succeeded! See if we are admin or user
		u, err := o.User()
		if err == nil {
			if u.Admin {
				joinpermissions = "admin"
			} else {
				joinpermissions = "user"
			}
		}

	}
	if !config.Get().Permissions[joinpermissions].Join {
		return errors.New("Joining is currently disabled.")
	}

	cfg := config.Get()
	if cfg.MaxUsers >= 0 {
		unum, err := Database.Userdb.CountUsers()
		if err != nil {
			return err
		}
		if uint64(cfg.MaxUsers) <= unum {
			return errors.New("The maximum number of users has been reached.")
		}
	}

	// Joining is allowed
	return nil
}

// JoinHandleGET handles joining ConnectorDB - the frontend of joining (ie GET)
func JoinHandleGET(writer http.ResponseWriter, request *http.Request) {
	err := checkIfJoinAllowed(request)
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	writer.WriteHeader(http.StatusOK)
	WWWJoin.Execute(writer, map[string]interface{}{
		"Version": connectordb.Version,
		"Join":    err == nil,
		"ErrMsg":  msg,
	})
}

// JoinHandlePOST handles the actual user creation based upon the given structure
func JoinHandlePOST(writer http.ResponseWriter, request *http.Request) {
	//
	writer.WriteHeader(http.StatusOK)
	WWWJoin.Execute(writer, map[string]interface{}{"Version": connectordb.Version})
}
