package crud

import (
	"connectordb/streamdb/operator"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"

	"connectordb/plugins/rest/restcore"
)

func getDevicePath(request *http.Request) (username string, devicename string, devicepath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	devicepath = username + "/" + devicename
	return username, devicename, devicepath
}

//ListDevices lists the devices that the given user has
func ListDevices(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname := mux.Vars(request)["user"]
	d, err := o.ReadAllDevices(usrname)
	return restcore.JSONWriter(writer, d, logger, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, devname, devpath := getDevicePath(request)
	err := restcore.ValidName(devname, nil)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	if err = o.CreateDevice(devpath); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	return ReadDevice(o, writer, request, logger)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)

	if err := restcore.BadQ(o, writer, request, logger); err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	d, err := o.ReadDevice(devpath)
	return restcore.JSONWriter(writer, d, logger, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)

	d, err := o.ReadDevice(devpath)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	err = restcore.UnmarshalRequest(request, d)
	err = restcore.ValidName(d.Name, err)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	if d.ApiKey == "" {
		//The user wants to reset the API key
		newkey, err := uuid.NewV4()
		if err != nil {
			return restcore.WriteError(writer, logger, http.StatusInternalServerError, err, false)
		}
		d.ApiKey = newkey.String()
	}

	if err = o.UpdateDevice(d); err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	return restcore.JSONWriter(writer, d, logger, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, devpath := getDevicePath(request)
	err := o.DeleteDevice(devpath)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}
	restcore.OK(writer)
	return 0, ""
}
