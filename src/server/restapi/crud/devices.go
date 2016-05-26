/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package crud

import (
	"connectordb/authoperator"
	"connectordb/users"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"

	"server/restapi/restcore"
	"server/webcore"
)

func getDevicePath(request *http.Request) (username string, devicename string, devicepath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	devicepath = username + "/" + devicename
	return username, devicename, devicepath
}

//ListDevices lists the devices that the given user has
func ListDevices(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]
	d, err := o.ReadUserDevicesToMap(usrname)
	return restcore.JSONWriter(writer, d, logger, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, devname, devpath := getDevicePath(request)
	err := restcore.ValidName(devname, nil)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	var dm users.DeviceMaker
	err = restcore.UnmarshalRequest(request, &dm)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	dm.Name = devname
	if err = o.CreateDevice(devpath, &dm); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	return ReadDevice(o, writer, request, logger)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	d, err := o.ReadDeviceToMap(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)

	var updates map[string]interface{}

	err := restcore.UnmarshalRequest(request, &updates)

	if err = o.UpdateDevice(devpath, updates); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	d, err := o.ReadDevice(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)
	err := o.DeleteDevice(devpath)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	restcore.OK(writer)
	return webcore.DEBUG, ""
}
