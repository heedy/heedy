/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"connectordb"
	"connectordb/operator"
	"connectordb/users"
	"net/http"
	"server/webcore"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

//TemplateData is the struct that is passed to the templates
type TemplateData struct {
	//These are information about the device performing the query
	ThisUser   *users.User
	ThisDevice *users.Device

	//This is info about the u/d/s that is being queried
	User   *users.User
	Device *users.Device
	Stream *users.Stream

	//When given a user or device, the user's Devices and device's Streams
	// are also exposed. When giving Index,
	//	both the current user's devices and current user's user device's streams
	//	are sent
	Devices []users.Device
	Streams []users.Stream

	//And some extra status info
	Status string
	Ref    string

	//The Database Version
	Version string
}

//GetTemplateData initializes the template
func GetTemplateData(o operator.Operator) (*TemplateData, error) {
	thisU, err := o.User()
	if err != nil {
		return nil, err
	}
	thisD, err := o.Device()
	return &TemplateData{
		ThisUser:   thisU,
		ThisDevice: thisD,
		Version:    connectordb.Version,
	}, err
}

//Index reads the index
func Index(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	td.Devices, err = o.ReadAllDevicesByUserID(td.ThisUser.UserId)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	td.Streams, err = o.ReadAllStreamsByDeviceID(td.ThisDevice.DeviceId)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppIndex.Execute(writer, td)
	return webcore.DEBUG, ""
}

//User reads the given user
func User(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	td.User, err = o.ReadUser(mux.Vars(request)["user"])
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}
	td.Devices, err = o.ReadAllDevicesByUserID(td.User.UserId)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}

	writer.WriteHeader(http.StatusOK)
	AppUser.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Device reads the given device
func Device(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	usr := mux.Vars(request)["user"]
	td.User, err = o.ReadUser(usr)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}
	dev := usr + "/" + mux.Vars(request)["device"]
	td.Device, err = o.ReadDevice(dev)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}
	td.Streams, err = o.ReadAllStreamsByDeviceID(td.Device.DeviceId)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}

	writer.WriteHeader(http.StatusOK)
	AppDevice.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Stream reads the given stream
func Stream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	usr := mux.Vars(request)["user"]
	td.User, err = o.ReadUser(usr)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}
	dev := usr + "/" + mux.Vars(request)["device"]
	td.Device, err = o.ReadDevice(dev)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}
	strm := dev + "/" + mux.Vars(request)["stream"]
	td.Stream, err = o.ReadStream(strm)
	if err != nil {
		return LoggedIn404(o, writer, logger, err)
	}

	writer.WriteHeader(http.StatusOK)
	AppStream.Execute(writer, td)
	return webcore.DEBUG, ""
}
