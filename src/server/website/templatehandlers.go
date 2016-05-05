/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package website

import (
	"connectordb"
	"connectordb/authoperator"
	"connectordb/users"
	"html/template"
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

	//And some extra status info
	Status string
	Ref    string

	//The Database Version
	Version string

	// The operator that this TemplateData uses.
	operator *authoperator.AuthOperator
}

func (td *TemplateData) DataURIToAttr(uri string) template.HTMLAttr {
	return template.HTMLAttr("src=\"" + uri + "\"")
}

//GetTemplateData initializes the template
func GetTemplateData(o *authoperator.AuthOperator, request *http.Request) (*TemplateData, error) {
	thisU, err := o.User()
	if err != nil {
		return nil, err
	}

	thisD, err := o.Device()
	if err != nil {
		return nil, err
	}

	// Partially construct the data
	td := &TemplateData{
		ThisUser:   thisU,
		ThisDevice: thisD,
		Version:    connectordb.Version,
		operator:   o,
	}

	// Now grab the session vars if they exist
	var usr, dev, stream string
	var ok bool

	if usr, ok = mux.Vars(request)["user"]; ok {
		td.User, err = o.ReadUser(usr)
		if err != nil {
			return nil, err
		}
	}

	if dev, ok = mux.Vars(request)["device"]; ok {
		dev = usr + "/" + dev

		td.Device, err = o.ReadDevice(dev)
		if err != nil {
			return nil, err
		}
	}

	if stream, ok = mux.Vars(request)["stream"]; ok {
		stream = dev + "/" + stream

		td.Stream, err = o.ReadStream(stream)
		if err != nil {
			return nil, err
		}
	}

	return td, err
}

// Reads the devices for the user requesting the page
func (t *TemplateData) ReadMyDevices() (out []*users.Device, err error) {
	return t.operator.ReadAllDevicesByUserID(t.ThisUser.UserID)
}

// Reads the streams for the user requesting the page
func (t *TemplateData) ReadMyStreams() (out []*users.Stream, err error) {
	return t.operator.ReadAllStreamsByDeviceID(t.ThisDevice.DeviceID)
}

// Reads the devices for the page's user
func (t *TemplateData) ReadDevices() (out []*users.Device, err error) {
	return t.operator.ReadAllDevicesByUserID(t.User.UserID)
}

// Reads the streams for the page's device
func (t *TemplateData) ReadStreams() (out []*users.Stream, err error) {
	return t.operator.ReadAllStreamsByDeviceID(t.Device.DeviceID)
}

// Reads all users on the system
func (t *TemplateData) ReadUsers() (out []*users.User, err error) {
	return t.operator.ReadAllUsers()
}

//Index reads the index
func Index(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	if o.Name() == "nobody" {
		// Nobody does not have access to the logged in index page
		return -1, ""
	}
	td, err := GetTemplateData(o, request)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppIndex.Execute(writer, td)
	return webcore.DEBUG, ""
}

//User reads the given user
func User(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o, request)
	if err != nil {
		if o.Name() == "nobody" {
			// Backtrack - show the nobody their login page
			return -1, ""
		}
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppUser.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Device reads the given device
func Device(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o, request)
	if err != nil {
		if o.Name() == "nobody" {
			// Backtrack - show the nobody their login page
			return -1, ""
		}
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppDevice.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Stream reads the given stream
func Stream(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o, request)
	if err != nil {
		if o.Name() == "nobody" {
			// Backtrack - show the nobody their login page
			return -1, ""
		}
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppStream.Execute(writer, td)
	return webcore.DEBUG, ""
}
