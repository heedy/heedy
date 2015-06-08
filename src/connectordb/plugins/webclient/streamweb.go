package webclient

import (
	"connectordb/streamdb/operator"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
)

const (
	defaultTemplate = `
{
	"type": "object",
	"properties": {
		"value": {
			"type": "number",
			"description":"A numeric value"
		}
	}
}`
)

func readStreamPage(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "ReadStreamPage")

	pageData := make(map[string]interface{})
	vars := mux.Vars(se.Request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, err := se.Operator.ReadStreamByID(int64(streamid))
	if err != nil {
		logger.Warn(err.Error())
		pageData["alert"] = "Error getting stream."
	}

	device, err := se.Operator.ReadDeviceByID(stream.DeviceId)
	if err != nil {
		logger.Warn(err.Error())
		pageData["alert"] = "Error getting stream."
	}
	logger.Debugf("Reading stream: %v/%v", device.Name, stream.Name)
	pageData["stream"] = stream
	pageData["user"] = se.User
	pageData["device"] = device
	pageData["flashes"] = se.Session.Flashes()

	se.Save()
	err = streamReadTemplate.ExecuteTemplate(se.Writer, "stream.html", pageData)
	if err != nil {
		logger.Error(err.Error())
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func editStreamAction(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "EditStreamAction")
	vars := mux.Vars(se.Request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, err := se.Operator.ReadStreamByID(int64(streamid))
	if err != nil {
		logger.Warn(err.Error())
		se.Session.AddFlash("Error getting stream, maybe it was deleted?")
		goto redirect
	}
	logger.Info("Update Stream: %v", stream.Name)
	err = se.Operator.UpdateStream(stream)

	if err != nil {
		logger.Warn(err.Error())
		se.Session.AddFlash(err.Error())
	} else {
		se.Session.AddFlash("Stream modified")
	}

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/stream/"+streamids, http.StatusTemporaryRedirect)
}

func createStreamAction(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "CreateStreamAction")

	vars := mux.Vars(se.Request)
	devids := vars["id"]
	streamtype := se.Request.PostFormValue("datatype")
	name := se.Request.PostFormValue("name")
	logger.Infof("Creating: %v", name)

	if streamtype == "" {
		streamtype = defaultTemplate
	}

	devid, _ := strconv.Atoi(devids)
	device, err := se.Operator.ReadDeviceByID(int64(devid))

	if err != nil {
		logger.Warn(err.Error())
		se.Session.AddFlash("Error getting device, maybe it was deleted?")
		goto redirect
	}

	err = se.Operator.CreateStreamByDeviceID(device.DeviceId, name, streamtype)

	if err != nil {
		logger.Warn(err.Error())
		se.Session.AddFlash("Error creating stream.")
	}

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
}

func insertStreamAction(se *SessionEnvironment, logger *log.Entry) {
	// Don't log anything user-related, as this may kill a little bit of security

	// Init all variables and check them
	vars := mux.Vars(se.Request)
	streamids := vars["id"]
	var datapointjson interface{}
	var datapoint operator.Datapoint
	streamData := se.Request.PostFormValue("formdata")

	streamid, err := strconv.Atoi(streamids)
	if se.HandleError(err, "Error inserting data; invalid stream id.", logger) {
		goto redirect
	}

	// Convert the data we need to json
	err = json.Unmarshal([]byte(streamData), &datapointjson)
	if se.HandleError(err, "Error inserting data; invalid format.", logger) {
		goto redirect
	}

	// Save the data
	datapoint = operator.NewDatapoint(datapointjson)
	err = se.Operator.InsertStreamByID(int64(streamid), []operator.Datapoint{datapoint}, "")
	if se.HandleError(err, "Error saving data.", logger) {
		goto redirect
	}

	se.Session.AddFlash("Created data point.")

redirect:
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/stream/"+streamids, http.StatusTemporaryRedirect)

}
