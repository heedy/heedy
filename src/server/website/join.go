/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"config"
	"connectordb"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/restapi/restcore"
	"server/webcore"
)

type recaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:error-codes`
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
	logger := webcore.GetRequestLogger(request, "join")
	err := checkIfJoinAllowed(request)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	cfg := config.Get()

	logger.Debug(msg)

	writer.WriteHeader(http.StatusOK)
	WWWJoin.Execute(writer, map[string]interface{}{
		"Version": connectordb.Version,
		"Join":    err == nil,
		"ErrMsg":  msg,
		"Captcha": cfg.Captcha.Enabled,
		"SiteKey": cfg.Captcha.SiteKey,
	})
}

type JoinStream struct {
	Name        string
	Nickname    string
	Description string
	Icon        string
	Schema      string
}

type JoinDevice struct {
}

type JoinUser struct {
	Name     string
	Password string
	Nickname string
}

type Joiner struct {
	Captcha string `json:"captcha"`
	Name    string `json:"name"`
}

// JoinHandlePOST handles the actual user creation based upon the given structure
func JoinHandlePOST(writer http.ResponseWriter, request *http.Request) {
	var j Joiner
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
	logger.Info("Join Succeeded")
	restcore.OK(writer)
}
