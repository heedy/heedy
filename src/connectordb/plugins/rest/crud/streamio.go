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
	return 0, querylog
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
			return 0, ""
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
	return 0, ""
}

//StreamRange gets a range of data from a stream
func StreamRange(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	_, _, _, streampath := restcore.GetStreamPath(request)
	logger = logger.WithField("op", "StreamRange")
	q := request.URL.Query()

	i1s := q.Get("i1")
	i2s := q.Get("i2")

	//If either i1 or i2 are given, then it is a range by index
	if len(i1s) > 0 || len(i2s) > 0 {
		i1, err := strconv.ParseInt(i1s, 0, 64)
		if i1s != "" && err != nil {
			return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		}

		i2, err := strconv.ParseInt(i2s, 0, 64)
		if i2s != "" && err != nil {
			return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		} else if i2s == "" {
			i2 = 0
			i2s = "Inf"
		}

		querylog := fmt.Sprintf("irange [%s,%s)", i1s, i2s)

		dr, err := o.GetStreamIndexRange(streampath, i1, i2)

		lvl, _ := writeJSONResult(writer, dr, logger, err)
		return lvl, querylog
	}

	//It is not a range by index. See if it is a range by time
	t1s := q.Get("t1")
	t2s := q.Get("t2")
	if len(t1s) > 0 || len(t2s) > 0 {
		t1, err := strconv.ParseFloat(t1s, 64)
		if err != nil {
			return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		}

		t2, err := strconv.ParseFloat(t2s, 64)
		if t2s != "" && err != nil {
			return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		} else if t2s == "" {
			t2 = 0.
			t2s = "Inf"
		}

		lims := q.Get("limit")
		lim, err := strconv.ParseUint(lims, 0, 64)
		if lims != "" && err != nil {
			return restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
		} else if lims == "" {
			lim = 0
			lims = "Inf"
		}

		querylog := fmt.Sprintf("trange [%s,%s) limit=%s", t1s, t2s, lims)
		dr, err := o.GetStreamTimeRange(streampath, t1, t2, int64(lim))
		if err != nil {
			lvl, _ := restcore.WriteError(writer, logger, http.StatusBadRequest, err, false)
			return lvl, querylog
		}
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
