package webclient

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

var (
	DefaultSchema = `{"type": "number"}`
	)

func readStreamPage(se *SessionEnvironment) {
	pageData := make(map[string]interface{})
	vars := mux.Vars(se.Request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, err := se.Operator.ReadStreamByID(int64(streamid))
	if err != nil {
		pageData["alert"] = "Error getting stream."
	}
	device, err := se.Operator.ReadDeviceByID(stream.DeviceId)
	if err != nil {
		pageData["alert"] = "Error getting stream."
	}

	pageData["stream"] = stream
	pageData["user"] = se.User
	pageData["device"] = device
	pageData["flashes"] = se.Session.Flashes()

	se.Save()
	err = streamReadTemplate.ExecuteTemplate(se.Writer, "stream.html", pageData)
	if err != nil {
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func editStreamAction(se *SessionEnvironment) {
	vars := mux.Vars(se.Request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, err := se.Operator.ReadStreamByID(int64(streamid))
	if err != nil {
		se.Session.AddFlash("Error getting stream, maybe it was deleted?")
		goto redirect
	}

	err = se.Operator.UpdateStream(stream)

	if err != nil {
		log.Errorf(err.Error())
		se.Session.AddFlash(err.Error())
	} else {
		se.Session.AddFlash("Created Device")
	}

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/stream/"+streamids, http.StatusTemporaryRedirect)
}

func createStreamAction(se *SessionEnvironment) {
	vars := mux.Vars(se.Request)
	devids := vars["id"]
	name := se.Request.PostFormValue("name")

	devid, _ := strconv.Atoi(devids)
	device, err := se.Operator.ReadDeviceByID(int64(devid))

	if err != nil {
		log.Errorf(err.Error())
		se.Session.AddFlash("Error getting device, maybe it was deleted?")
		goto redirect
	}

	err = se.Operator.CreateStreamByDeviceID(device.DeviceId, name, DefaultSchema)

	if err != nil {
		log.Errorf(err.Error())
		se.Session.AddFlash("Error creating stream.")
	}

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
}
