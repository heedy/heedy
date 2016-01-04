/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"config"
	"connectordb"
	"connectordb/operator"
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

func checkIfJoinAllowed(request *http.Request) error {
	if !webcore.IsActive {
		return errors.New("ConnectorDB is currently disabled.")
	}
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
	cfg := config.Get()
	if !cfg.Permissions[joinpermissions].Join {
		return errors.New(cfg.Permissions[joinpermissions].JoinDisabledMessage)
	}

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
	tstart := time.Now()
	logger := webcore.GetRequestLogger(request, "join")
	err := checkIfJoinAllowed(request)
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

type JoinStream struct {
	Name        string      `json:"name"`
	Nickname    string      `json:"nickname"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
	Schema      interface{} `json:"schema"`
}

type Joiner struct {
	Captcha  string       `json:"captcha"`
	Name     string       `json:"name"`
	Nickname string       `json:"nickname"`
	Email    string       `json:"email"`
	Password string       `json:"password"`
	Icon     string       `json:"icon"`
	Streams  []JoinStream `json:"streams"`
}

// JoinHandlePOST handles the actual user creation based upon the given structure
func JoinHandlePOST(writer http.ResponseWriter, request *http.Request) {

	tstart := time.Now()

	var j Joiner
	var schema []byte
	var usr *users.User
	var dev *users.Device
	var strm *users.Stream
	var uo operator.Operator
	logger := webcore.GetRequestLogger(request, "JOIN")

	// First check if join is allowed at all
	err := checkIfJoinAllowed(request)
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
	db := operator.NewOperator(Database)
	// OK - now set up the user
	err = db.CreateUser(j.Name, j.Email, j.Password)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return
	}

	// Now update the user with nickname/icon

	usr, err = db.ReadUser(j.Name)
	if err != nil {
		goto errfail
	}
	usr.Nickname = j.Nickname
	usr.Icon = j.Icon
	err = db.UpdateUser(usr)
	if err != nil {
		db.DeleteUser(j.Name)
		restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		return
	}

	// Now create the streams using the user's operator
	uo, err = operator.NewUserOperator(Database, j.Name)

	// Now create the streams
	dev, err = uo.ReadDeviceByUserID(usr.UserId, "user")
	if err != nil {
		goto errfail
	}
	for i := range j.Streams {
		schema, err = json.Marshal(j.Streams[i].Schema)
		if err != nil {
			goto errfail
		}

		err = uo.CreateStreamByDeviceID(dev.DeviceId, j.Streams[i].Name, string(schema))
		if err != nil {
			goto errfail
		}

		// Now update the stream with the extra values
		strm, err = uo.ReadStreamByDeviceID(dev.DeviceId, j.Streams[i].Name)
		if err != nil {
			goto errfail
		}

		strm.Nickname = j.Streams[i].Nickname
		strm.Icon = j.Streams[i].Icon
		strm.Description = j.Streams[i].Description

		err = uo.UpdateStream(strm)
		if err != nil {
			goto errfail
		}
	}

	// Great success! The user was created successfully. We now write the cookie for the user
	webcore.CreateSessionCookie(uo, writer, request)
	webcore.LogRequest(logger, webcore.INFO, fmt.Sprintf("User '%s' Joined", j.Name), time.Since(tstart))
	restcore.OK(writer)
	return

errfail:
	db.DeleteUser(j.Name)
	restcore.WriteError(writer, logger, http.StatusInternalServerError, err, false)
	webcore.LogRequest(logger, webcore.WARNING, "", time.Since(tstart))
	return
}
