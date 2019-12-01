package run

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

type StartMessage struct {
	API string `json:"api"`
}

type APIHandler struct {
	Log *logrus.Entry
	M   *Manager
	H   http.Handler
	V   assets.RunType
}

func NewAPIHandler(M *Manager, runtype string, rtv assets.RunType) *APIHandler {
	l := logrus.WithField("runtype", runtype)
	return &APIHandler{
		M:   M,
		Log: l,
		V:   rtv,
	}
}

func (ah *APIHandler) setup() error {
	if ah.H == nil {
		h, err := ah.M.GetHandler("", *ah.V.API)
		if err != nil {
			return err
		}
		ah.H = h
	}
	return nil
}

func (ah *APIHandler) Start(i *Info) (http.Handler, error) {
	if err := ah.setup(); err != nil {
		return nil, err
	}
	if ah.M.DB.Verbose {
		logrus.Debugf("POST %s", *ah.V.API)
	}
	b, err := Request(ah.H, "POST", "", i, nil)
	if err != nil {
		return nil, err
	}
	var sm StartMessage
	err = json.Unmarshal(b.Bytes(), &sm)
	if err != nil {
		ah.Kill(i.APIKey)
		return nil, err
	}
	var h http.Handler
	if sm.API != "" {
		hv, err := NewReverseProxy(ah.M.DB.Assets().DataDir(), sm.API)
		if err != nil {
			ah.Kill(i.APIKey)
			return nil, err
		}
		h = hv
		method, host, err := GetEndpoint(ah.M.DB.Assets().DataDir(), sm.API)
		if err != nil {
			ah.Kill(i.APIKey)
			return nil, err
		}
		if err = WaitForAPI(method, host, 30*time.Second); err != nil {
			ah.Kill(i.APIKey)
			return nil, err
		}
	}

	return h, nil
}

func (ah *APIHandler) Run(i *Info) error {
	if err := ah.setup(); err != nil {
		return err
	}
	if ah.M.DB.Verbose {
		logrus.Debugf("PATCH %s", *ah.V.API)
	}
	_, err := Request(ah.H, "PATCH", "", i, nil)
	return err
}

func (ah *APIHandler) Stop(apikey string) error {
	if err := ah.setup(); err != nil {
		return err
	}
	if ah.M.DB.Verbose {
		logrus.Debugf("DELETE %s", singleJoiningSlash(*ah.V.API, url.PathEscape(apikey)))
	}
	_, err := Request(ah.H, "DELETE", "/"+url.PathEscape(apikey), nil, nil)
	return err
}

func (ah *APIHandler) Kill(apikey string) error {
	if err := ah.setup(); err != nil {
		return err
	}
	if ah.M.DB.Verbose {
		logrus.Debugf("DELETE %s?kill=true", singleJoiningSlash(*ah.V.API, url.PathEscape(apikey)))
	}
	_, err := Request(ah.H, "DELETE", "/"+url.PathEscape(apikey)+"?kill=true", nil, nil)
	return err
}
