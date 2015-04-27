package webclient

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)


func createDeviceAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	user, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	devname := request.PostFormValue("name")

	log.Printf("Creating device %v", devname)

	err := operator.CreateDevice(devname, user)
	if err != nil {
		session.AddFlash("You must enter a device name that isn't empty or taken.")
	} else {
		session.AddFlash("Created Device")
	}

	http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
}

func editDevicePage(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	_, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	vars := mux.Vars(request)
	devids := vars["id"]
	devid, _ := strconv.Atoi(devids)
	device, err := userdb.ReadDeviceById(int64(devid))

	origDevice := *device

	if err != nil {
		session.AddFlash("Error getting device, maybe it was deleted?")
		goto redirect
	}

	device.Nickname = request.PostFormValue("shortname")
	device.Enabled = request.PostFormValue("enabled") == "checked"
	device.IsAdmin = request.PostFormValue("superdevice") == "checked"
	device.CanWrite = request.PostFormValue("canwrite") == "checked"
	device.CanWriteAnywhere = request.PostFormValue("canwriteanywhere") == "checked"
	device.CanActAsUser = request.PostFormValue("userproxy") == "checked"

	err = operator.UpdateDevice(device, &origDevice)

	if err != nil {
		log.Printf(err.Error())
		session.AddFlash(err.Error())
	} else {
		session.AddFlash("Created Device")
	}

redirect:
	http.Redirect(writer, request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
}


func getDevicePage(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	//user, userdevice, _ := srw.GetUserAndDevice()
	user, _, _ := srw.GetUserAndDevice()
	//operator, _ := userdb.GetOperatorForDevice(userdevice)
	pageData := make(map[string]interface{})

	vars := mux.Vars(request)
	devids := vars["id"]

	devid, _ := strconv.Atoi(devids)

	device, err := userdb.ReadDeviceById(int64(devid))
	pageData["device"] = device
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting device."
	}

	streams, err := userdb.ReadStreamsByDevice(device.DeviceId)
	pageData["streams"] = streams

	if err != nil {
		pageData["alert"] = "Error getting device streams"
	}

	err = deviceInfoTemplate.ExecuteTemplate(writer, "device_info.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
