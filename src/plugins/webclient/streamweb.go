package webclient

import (
	"log"
	"net/http"
	"strconv"


	"github.com/gorilla/mux"
)

func readStreamPage(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	user, userdevice, _ := srw.GetUserAndDevice()
	//user, _, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)
	pageData := make(map[string]interface{})

	vars := mux.Vars(request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, device, err := operator.ReadStreamById(int64(streamid))

	if err != nil {
		pageData["alert"] = "Error getting stream."
	}

	pageData["stream"] = stream
	pageData["user"] = user
	pageData["device"] = device
	pageData["flashes"] = session.Flashes()

	err = streamReadTemplate.ExecuteTemplate(writer, "stream.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func editStreamAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	_, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	vars := mux.Vars(request)
	streamids := vars["id"]
	streamid, _ := strconv.Atoi(streamids)
	stream, device, err := operator.ReadStreamById(int64(streamid))

	origstream := *stream

	if err != nil {
		session.AddFlash("Error getting stream, maybe it was deleted?")
		goto redirect
	}


	err = operator.UpdateStream(device, stream, &origstream)

	if err != nil {
		log.Printf(err.Error())
		session.AddFlash(err.Error())
	} else {
		session.AddFlash("Created Device")
	}

redirect:
	http.Redirect(writer, request, "/secure/stream/" + streamids, http.StatusTemporaryRedirect)
}

func createStreamAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	_, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	vars := mux.Vars(request)
	devids := vars["id"]

	devid, _ := strconv.Atoi(devids)
	device, err := operator.ReadDeviceById(int64(devid))

	if err != nil {
		log.Printf(err.Error())
		session.AddFlash("Error getting device, maybe it was deleted?")
		http.Redirect(writer, request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
	}

	name := request.PostFormValue("name")
	err = operator.CreateStream(name, "x", device)

	if err != nil {
		log.Printf(err.Error())
		session.AddFlash("Error creating stream.")
	}

	http.Redirect(writer, request, "/secure/device/"+devids, http.StatusTemporaryRedirect)
}
