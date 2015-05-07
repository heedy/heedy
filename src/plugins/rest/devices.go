package rest

import (
	"log"
	"net/http"
	"streamdb"
	"strings"

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
	//Don't log, since this is just a ping - we don't want to spam the logs
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
	}
}

//ListDevices lists the devices that the given user has
func ListDevices(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	usrname := strings.ToLower(mux.Vars(request)["user"])
	log.Printf("%s: List Devices for %s\n", o.Name(), usrname)
	d, err := o.ReadAllDevices(usrname)
	return JSONWriter(writer, d, err)
}

//CreateDevice creates a new user from a REST API request
func CreateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, devname, devpath := getDevicePath(request)
	log.Printf("%s: Create Device %s\n", o.Name(), devpath)
	err := ValidName(devname, nil)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return err
	}

	if err = o.CreateDevice(devpath); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}

	return ReadDevice(o, writer, request)
}

//ReadDevice gets an existing device from a REST API request
func ReadDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	log.Printf("%s: Read Device %s\n", o.Name(), devpath)
	d, err := o.ReadDevice(devpath)

	return JSONWriter(writer, d, err)
}

//UpdateDevice updates the metadata for existing device from a REST API request
func UpdateDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	log.Printf("%s: Update Device %s\n", o.Name(), devpath)

	d, err := o.ReadDevice(devpath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}

	err = UnmarshalRequest(request, d)
	err = ValidName(d.Name, err)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return err
	}

	if d.ApiKey == "" {
		//The user wants to reset the API key
		newkey, err := uuid.NewV4()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return err
		}
		d.ApiKey = newkey.String()
	}

	if err = o.UpdateDevice(devpath, d); err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}
	return JSONWriter(writer, d, err)
}

//DeleteDevice deletes existing device from a REST API request
func DeleteDevice(o streamdb.Operator, writer http.ResponseWriter, request *http.Request) error {
	_, _, devpath := getDevicePath(request)
	log.Printf("%s: Delete Device %s\n", o.Name(), devpath)
	err := o.DeleteDevice(devpath)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return err
	}
	return OK(writer)
}
