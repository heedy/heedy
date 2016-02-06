/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package webcore

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

//Log levels supported by LogRequest
const (
	DEBUG   = iota
	INFO    = iota
	WARNING = iota
	ERROR   = iota
)

//GetRequestLogger returns a logrus log entry which has fields prepopulated for the request
func GetRequestLogger(request *http.Request, opname string) *log.Entry {
	//Since an important use case is behind nginx, the following rule is followed:
	//localhost address is not logged if real-ip header exists (since it is from localhost)
	//if real-ip header exists, faddr=address (forwardedAddress) is logged
	//In essence, if behind nginx, there is no need for the addr=blah

	fields := log.Fields{"addr": request.RemoteAddr, "uri": request.URL.Path, "op": opname}
	if realIP := request.Header.Get("X-Real-IP"); realIP != "" {
		fields["faddr"] = realIP
		if strings.HasPrefix(request.RemoteAddr, "127.0.0.1") || strings.HasPrefix(request.RemoteAddr, "::1") {
			delete(fields, "addr")
		}
	}

	return log.WithFields(fields)
}

// LogRequest writes a log message given the log entry to use, a log level, optional text, and the query duration
func LogRequest(l *log.Entry, loglevel int, txt string, tdiff time.Duration) {
	//Set up how the log message is printed for this query
	if txt == "" {
		txt = tdiff.String()
	} else {
		txt += " - " + tdiff.String()
	}
	switch loglevel {
	case DEBUG:
		l.Debugln(txt)
	case INFO:
		l.Infoln(txt)
	case WARNING:
		l.Warningln(txt)
	case ERROR:
		l.Errorln(txt)
	}
}
