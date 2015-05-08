package rest

import (
	"net/http"
	"streamdb"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
)

func getDevicePath(request *http.Request) (username string, devicename string, devicepath string) {
	username = strings.ToLower(mux.Vars(request)["user"])
	devicename = strings.ToLower(mux.Vars(request)["device"])
	devicepath = username + "/" + devicename
	return username, devicename, devicepath
}

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "PingThis"}).Debugln()
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(o.Name()))
	return nil
}

//GetDevice handles a {user}/{device} request
func GetDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	devname := strings.ToLower(mux.Vars(request)["device"])

	switch devname {
	default:
		return ReadDevice(o, writer, request)
	case "ls":
		return ListDevices(o, writer, request)
	case "favicon.ico":
		writer.WriteHeader(http.StatusNotFound)
		log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr}).Debugln("Request for favicon")
		return nil
	}
}

//ListDevices lists the devices that the given user has
func ListDevices(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ListDevices", "arg": usrname})
	logger.Debugln()
	d, err := o.ReadAllDevices(usrname)
	return JSONWriter(writer, d, logger, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, devname, devpath := getDevicePath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "CreateDevice", "arg": devpath})
	logger.Infoln()
	err := ValidName(devname, nil)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}

	if err = o.CreateDevice(devpath); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	return ReadDevice(o, writer, request)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "ReadDevice", "arg": devpath})
	logger.Debugln()
	d, err := o.ReadDevice(devpath)

	return JSONWriter(writer, d, logger, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "UpdateDevice", "arg": devpath})
	logger.Infoln()

	d, err := o.ReadDevice(devpath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	err = UnmarshalRequest(request, d)
	err = ValidName(d.Name, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}

	if d.ApiKey == "" {
		//The user wants to reset the API key
		newkey, err := uuid.NewV4()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Errorln(err)
			return err
		}
		d.ApiKey = newkey.String()
	}

	if err = o.UpdateDevice(devpath, d); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return JSONWriter(writer, d, logger, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	logger := log.WithFields(log.Fields{"dev": o.Name(), "addr": request.RemoteAddr, "op": "DeleteDevice", "arg": devpath})
	logger.Infoln()
	err := o.DeleteDevice(devpath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return OK(writer)
}
