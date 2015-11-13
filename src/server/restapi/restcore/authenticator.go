package restcore

import (
	"connectordb"
	"net/http"
	"server/webcore"
	"sync/atomic"
	"time"
)

//Authenticator runs authentication on a request, making sure that the REST API can handle it
func Authenticator(apifunc webcore.APIHandler, db *connectordb.Database) http.HandlerFunc {
	funcname := webcore.GetFuncName(apifunc)
	qtimer := webcore.GetQueryTimer(funcname)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !webcore.IsActive {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte(`{"code": 503, "msg": "ConnectorDB is currently inactive", "ref": "DISABLED"}`))
			return
		}
		tstart := time.Now()

		atomic.AddUint32(&webcore.StatsRESTQueries, 1)
		atomic.AddInt32(&webcore.StatsActive, 1)
		defer atomic.AddInt32(&webcore.StatsActive, -1)

		logger := webcore.GetRequestLogger(request, funcname)

		//Handle a panic without crashing the whole rest interface
		defer webcore.HandlePanic(logger)

		webcore.WriteAccessControlHeaders(writer,request)

		o, err := webcore.Authenticate(db, request)
		if err != nil {
			WriteError(writer, logger, http.StatusUnauthorized, err, false)
			return
		}
		l := logger.WithField("dev", o.Name())

		//Alright, run the api function
		loglevel, txt := apifunc(o, writer, request, l)

		//Find the time that this query took
		tdiff := time.Since(tstart)
		qtimer.Add(tdiff)

		webcore.LogRequest(l, loglevel, txt, tdiff)

	})
}
