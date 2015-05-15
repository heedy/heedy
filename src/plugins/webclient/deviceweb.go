package webclient

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func createDeviceAction(se *SessionEnvironment) {
	devname := se.Request.PostFormValue("name")

	log.Printf("Creating device %v", devname)

	err := se.Operator.CreateDeviceByUserID(se.User.UserId, devname)
	if err != nil {
		se.Session.AddFlash("You must enter a device name that isn't empty or taken.")
	} else {
		se.Session.AddFlash("Created Device")
	}

	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/", http.StatusTemporaryRedirect)
}

func editDevicePage(se *SessionEnvironment) {
	vars := mux.Vars(se.Request)
	devids := vars["id"]
	devid, _ := strconv.Atoi(devids)
	device, err := se.Operator.ReadDeviceByID(int64(devid))

	if err != nil {
		se.Session.AddFlash("Error getting device, maybe it was deleted?")
		goto redirect
	}

	device.Nickname = se.Request.PostFormValue("shortname")
	device.Enabled = se.Request.PostFormValue("enabled") == "checked"
	device.IsAdmin = se.Request.PostFormValue("superdevice") == "checked"
	device.CanWrite = se.Request.PostFormValue("canwrite") == "checked"
	device.CanWriteAnywhere = se.Request.PostFormValue("canwriteanywhere") == "checked"
	device.CanActAsUser = se.Request.PostFormValue("userproxy") == "checked"

	err = se.Operator.UpdateDevice(device)

	if err != nil {
		log.Printf(err.Error())
		se.Session.AddFlash(err.Error())
	} else {
		se.Session.AddFlash("Updated Device")
	}

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
}

func getDevicePage(se *SessionEnvironment) {
	pageData := make(map[string]interface{})

	vars := mux.Vars(se.Request)
	devids := vars["id"]

	devid, _ := strconv.Atoi(devids)

	device, err := se.Operator.ReadDeviceByID(int64(devid))
	pageData["device"] = device
	pageData["user"] = se.User
	pageData["flashes"] = se.Session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting device."
	}

	streams, err := se.Operator.ReadAllStreamsByDeviceID(device.DeviceId)
	pageData["streams"] = streams

	if err != nil {
		pageData["alert"] = "Error getting device streams"
	}

	se.Save()
	err = deviceInfoTemplate.ExecuteTemplate(se.Writer, "device_info.html", pageData)
	if err != nil {
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}
