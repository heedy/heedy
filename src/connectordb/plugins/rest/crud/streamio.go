package crud

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrRangeArgs is thrown when invalid arguments are given to trange
	ErrRangeArgs = errors.New(`A range needs [both "i1" and "i2" int] or ["t1" and ["t2" decimal and/or "limit" int]]`)
	//ErrTime2IndexArgs is the error when args are incorrectly given to t2i
	ErrTime2IndexArgs = errors.New(`time2index requires an argument of "t" which is a decimal timestamp`)
)

//StreamLength gets the stream length
func StreamLength(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	l, err := o.LengthStream(streampath)

	return restcore.IntWriter(writer, l, logger, err)
}

//WriteStream writes the given stream
func WriteStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)

	var datapoints []datastream.Datapoint
	err := restcore.UnmarshalRequest(request, &datapoints)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}
	restamp := request.Method == "PUT"

	querylog := fmt.Sprintf("Insert %d", len(datapoints))
	if restamp {
		querylog += " (restamp)"
	}

	err = o.InsertStream(streampath, datapoints, restamp)
	if err != nil {
		lvl, _ := restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return lvl, querylog
	}
	atomic.AddUint32(&restcore.StatsInserts, uint32(len(datapoints)))
	restcore.OK(writer)
	return restcore.DEBUG, querylog
}

func writeJSONResult(writer http.ResponseWriter, dr datastream.DataRange, logger *log.Entry, err error) (int, string) {
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
	}

	jreader, err := operator.NewJsonReader(dr)
	if err != nil {
		if err == io.EOF {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.Header().Set("Content-Length", "2")
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("[]")) //If there are no datapoints, just return empty
			return restcore.DEBUG, ""
		}
		return restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
	}

	defer jreader.Close()
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(writer, jreader)
	if err != nil {
		logger.Errorln(err)
		return 3, err.Error()
	}
	return restcore.DEBUG, ""
}

//StreamRange gets a range of data from a stream
func StreamRange(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)
	q := request.URL.Query()

	i1, i2, err := restcore.ParseIRange(q)
	if err == nil {
		querylog := fmt.Sprintf("irange [%d,%d)", i1, i2)
		dr, err := o.GetStreamIndexRange(streampath, i1, i2)
		lvl, _ := writeJSONResult(writer, dr, logger, err)
		return lvl, querylog
	} else if err != restcore.ErrCantParse {
		return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
	}

	//The error is ErrCantParse - meaning that i1 and i2 are not present in query

	t1, t2, lim, err := restcore.ParseTRange(q)
	if err == nil {
		querylog := fmt.Sprintf("trange [%.1f,%.1f) limit=%d", t1, t2, lim)
		dr, err := o.GetStreamTimeRange(streampath, t1, t2, lim)
		lvl, _ := writeJSONResult(writer, dr, logger, err)
		return lvl, querylog
	}

	//None of the limits were recognized. Rather than exploding, return bad request
	return restcore.WriteError(writer, logger, http.StatusBadRequest, ErrRangeArgs, false)
}

//StreamTime2Index gets the time associated with the index
func StreamTime2Index(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)
	logger = logger.WithField("op", "Time2Index")

	ts := request.URL.Query().Get("t")
	t, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusForbidden, ErrTime2IndexArgs, false)
	}
	logger.Debugln("t=", ts)

	i, err := o.TimeToIndexStream(streampath, t)

	lvl, _ := restcore.JSONWriter(writer, i, logger, err)
	return lvl, "t=" + ts
}
