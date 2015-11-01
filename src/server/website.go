package server

import (
	"connectordb"
	"connectordb/operator"
	"connectordb/users"
	"errors"
	"net/http"
	"server/webcore"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"

	log "github.com/Sirupsen/logrus"
)

//WriteError writes the templated error page
func WriteError(logger *log.Entry, writer http.ResponseWriter, status int, err error, iserr bool) (int, string) {
	errmap := map[string]interface{}{
		"code": status,
		"msg":  err.Error(),
	}
	u, err2 := uuid.NewV4()
	if err2 != nil {
		logger.WithField("ref", "OSHIT").Errorln("Failed to generate error UUID: " + err2.Error())
		logger.WithField("ref", "OSHIT").Warningln("Original Error: " + err.Error())
		writer.WriteHeader(520)

		errmap["msg"] = "Failed to generate error UUID"
		errmap["ref"] = "OSHIT"
		return webcore.INFO, ""
	}
	errmap["ref"] = u.String()
	//Now that we have the error message, we log it and send the messages
	l := logger.WithFields(log.Fields{"ref": u.String(), "code": status})
	if iserr {
		l.Errorln(err.Error())
	} else {
		l.Warningln(err.Error())
	}

	writer.WriteHeader(status)
	AppError.Execute(writer, errmap)

	return webcore.INFO, ""
}

//Authenticator runs an auth check and either goes to the www template given or to the apifunc handler
func Authenticator(www *FileTemplate, apifunc webcore.APIHandler, db *connectordb.Database) http.HandlerFunc {
	funcname := webcore.GetFuncName(apifunc)
	qtimer := webcore.GetQueryTimer(funcname)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		usr, ok := mux.Vars(request)["user"]
		if ok && usr == "api" || usr == WWWPrefix || usr == AppPrefix {
			//We don't want to pass 404s to the handlers
			NotFoundHandler(writer, request)
			return
		}

		tstart := time.Now()
		logger := webcore.GetRequestLogger(request, funcname)

		//We don't need the "op" here
		delete(logger.Data, "op")

		//Handle a panic without crashing the whole rest interface
		defer webcore.HandlePanic(logger)

		webcore.WriteAccessControlHeaders(writer)

		if webcore.HasSession(request) {
			//There is a session associated with ConnectorDB

			if !webcore.IsActive {
				WriteError(logger, writer, http.StatusServiceUnavailable, errors.New("ConnectorDB app is disabled for maintenance."), false)
				return
			}

			atomic.AddInt32(&webcore.StatsActive, 1)
			defer atomic.AddInt32(&webcore.StatsActive, -1)

			o, err := webcore.Authenticate(db, request)
			if err == nil {
				//There is a user logged in
				l := logger.WithField("dev", o.Name())

				//Only count valid web queries
				atomic.AddUint32(&webcore.StatsWebQueries, 1)

				loglevel, txt := apifunc(o, writer, request, l)

				//Find the time that this query took
				tdiff := time.Since(tstart)
				qtimer.Add(tdiff)

				webcore.LogRequest(l, loglevel, txt, tdiff)
				return
			}

		}

		//If we got here, the user is not logged in. We therefore execute the "www" template given
		www.Execute(writer, map[string]string{
			"version": connectordb.Version,
		})

		webcore.LogRequest(logger, webcore.DEBUG, "", time.Since(tstart))
	})
}

//TemplateData is the struct that is passed to the templates
type TemplateData struct {
	//These are information about the device performing the query
	ThisUser   *users.User
	ThisDevice *users.Device

	//This is info about the u/d/s that is being queried
	User   *users.User
	Device *users.Device
	Stream *users.Stream
}

//GetTemplateData initializes the template
func GetTemplateData(o operator.Operator) (*TemplateData, error) {
	thisU, err := o.User()
	if err != nil {
		return nil, err
	}
	thisD, err := o.Device()
	return &TemplateData{
		ThisUser:   thisU,
		ThisDevice: thisD,
	}, nil
}

//Index reads the index
func Index(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppIndex.Execute(writer, td)
	return webcore.DEBUG, ""
}

//User reads the given user
func User(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	td.User, err = o.ReadUser(mux.Vars(request)["user"])
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppUser.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Device reads the given device
func Device(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	usr := mux.Vars(request)["user"]
	td.User, err = o.ReadUser(usr)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	dev := usr + "/" + mux.Vars(request)["device"]
	td.Device, err = o.ReadDevice(dev)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppDevice.Execute(writer, td)
	return webcore.DEBUG, ""
}

//Stream reads the given stream
func Stream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	td, err := GetTemplateData(o)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	usr := mux.Vars(request)["user"]
	td.User, err = o.ReadUser(usr)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	dev := usr + "/" + mux.Vars(request)["device"]
	td.Device, err = o.ReadDevice(dev)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}
	strm := dev + "/" + mux.Vars(request)["stream"]
	td.Stream, err = o.ReadStream(strm)
	if err != nil {
		return WriteError(logger, writer, http.StatusUnauthorized, err, false)
	}

	writer.WriteHeader(http.StatusOK)
	AppStream.Execute(writer, td)
	return webcore.DEBUG, ""
}
