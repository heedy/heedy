/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"config"
	"connectordb"
	"connectordb/authoperator"
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/restapi/restcore"
	"server/webcore"
	"time"

	pconfig "config/permissions"
)

type recaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// VerifyCaptcha checks if the captcha was solved successfully
func VerifyCaptcha(response string) (bool, error) {
	var rr recaptchaResponse
	if response == "" {
		return false, errors.New("No captcha response")
	}
	c := config.Get().Frontend.Captcha
	res, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {c.SiteSecret}, "response": {response}})
	if err != nil {
		return false, err
	}
	val, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(val, &rr)
	if err != nil {
		return false, err
	}
	return rr.Success, nil
}

func checkIfJoinAllowed(request *http.Request) (*pconfig.UserRole, error) {
	if !webcore.IsActive {
		return nil, errors.New("ConnectorDB is currently disabled.")
	}

	// First check if the user is authenticated (ie, a user is trying to add another user)
	o, err := webcore.Authenticate(Database, request)
	if err != nil {
		return nil, err
	}
	u, err := o.User()
	if err != nil {
		return nil, err
	}

	// We now have an operator which has join permissions - extract them!

	perm := pconfig.Get()
	r := permissions.GetUserRole(perm, u)
	if !r.Join {
		return r, errors.New(r.JoinDisabledMessage)
	}

	// Show a message if max users is reached
	if perm.MaxUsers >= 0 {
		unum, err := Database.Userdb.CountUsers()
		if err != nil {
			return nil, err
		}
		if perm.MaxUsers <= unum {
			return nil, errors.New("The maximum number of users has been reached.")
		}
	}

	// Joining is allowed
	return r, nil
}

// JoinHandleGET handles joining ConnectorDB - the frontend of joining (ie GET)
func JoinHandleGET(writer http.ResponseWriter, request *http.Request) {
	tstart := time.Now()
	logger := webcore.GetRequestLogger(request, "join")
	_, err := checkIfJoinAllowed(request)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	cfg := config.Get()

	writer.WriteHeader(http.StatusOK)
	WWWJoin.Execute(writer, map[string]interface{}{
		"Version": connectordb.Version,
		"Join":    err == nil,
		"ErrMsg":  msg,
		"Captcha": cfg.Captcha.Enabled,
		"SiteKey": cfg.Captcha.SiteKey,
	})
	webcore.LogRequest(logger, webcore.DEBUG, msg, time.Since(tstart))
}

// JoinStream is the structure used to encode a stream used for join
type JoinStream struct {
	Name        string      `json:"name"`
	Nickname    string      `json:"nickname"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
	Schema      interface{} `json:"schema"`
}

// Joiner is the struct sent in when POST to join, which creates the desired user structure.
type Joiner struct {
	users.UserMaker
	Captcha string `json:"captcha"`
}

// JoinHandlePOST handles the actual user creation based upon the given structure
// The only difference here than CRUD CreateUser is that the user creation is done with the
// admin operator
func JoinHandlePOST(writer http.ResponseWriter, request *http.Request) {

	tstart := time.Now()

	var j Joiner
	var uo *authoperator.AuthOperator
	logger := webcore.GetRequestLogger(request, "JOIN")

	// First check if join is allowed at all
	role, err := checkIfJoinAllowed(request)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return
	}

	// Get the request
	err = restcore.UnmarshalRequest(request, &j)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return
	}

	cfg := config.Get()

	if cfg.Frontend.Captcha.Enabled {
		verifyResult, err := VerifyCaptcha(j.Captcha)
		if !verifyResult {
			if err == nil {
				err = errors.New("Captcha Failed")
			}
			restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
			return
		}
	}
	// OK - now set up the user
	j.UserMaker.Role = role.JoinRole
	err = Database.CreateUser(&j.UserMaker)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return
	}

	// Great success! The user was created successfully. We now write the cookie for the user
	webcore.CreateSessionCookie(uo, writer, request)
	webcore.LogRequest(logger, webcore.INFO, fmt.Sprintf("User '%s' Joined", j.Name), time.Since(tstart))
	restcore.OK(writer)
	return
}
