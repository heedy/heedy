package rest

import (
	"connectordb/streamdb/operator"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
)

func getDevicePath(request *http.Request) (username string, devicename string, devicepath string) {
	username = mux.Vars(request)["user"]
	devicename = mux.Vars(request)["device"]
	devicepath = username + "/" + devicename
	return username, devicename, devicepath
}

//GetThis is a command to return the "username/devicename" of the currently authenticated thing
func GetThis(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	logger.WithField("op", "PingThis").Debugln()
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(o.Name()))
	return nil
}

//ListDevices lists the devices that the given user has
func ListDevices(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	logger = logger.WithField("op", "ListDevices")
	logger.Debugln()
	usrname := mux.Vars(request)["user"]
	d, err := o.ReadAllDevices(usrname)
	return JSONWriter(writer, d, logger, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	logger = logger.WithField("op", "CreateDevice")
	logger.Infoln()
	_, devname, devpath := getDevicePath(request)
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

	return ReadDevice(o, writer, request, logger)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)

	if err := BadQ(o, writer, request, logger); err != nil {
		return err
	}

	logger = logger.WithField("op", "ReadDevice")
	logger.Debugln()
	d, err := o.ReadDevice(devpath)

	return JSONWriter(writer, d, logger, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)
	logger = logger.WithField("op", "UpdateDevice")
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

	if err = o.UpdateDevice(d); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return JSONWriter(writer, d, logger, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, devpath := getDevicePath(request)
	logger = logger.WithField("op", "DeleteDevice")
	logger.Infoln()
	err := o.DeleteDevice(devpath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}
	return OK(writer)
}
