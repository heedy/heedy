package rest

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator"
	"errors"
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

//GetStreamLength gets the stream length
func GetStreamLength(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "StreamLength")
	logger.Debugln()

	l, err := o.LengthStream(streampath)

	return JSONWriter(writer, l, logger, err)
}

//WriteStream writes the given stream
func WriteStream(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "WriteStream")

	var datapoints []datastream.Datapoint
	err := UnmarshalRequest(request, &datapoints)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}
	restamp := request.Method == "PATCH"

	logger.Debugln("Inserting", len(datapoints), "dp restamp=", restamp)

	err = o.InsertStream(streampath, datapoints, restamp)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln(err)
		return err
	}
	atomic.AddUint32(&StatsInserts, uint32(len(datapoints)))
	return OK(writer)
}

func writeJSONResult(writer http.ResponseWriter, dr datastream.DataRange, logger *log.Entry, err error) error {
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		logger.Warningln(err)
		return err
	}

	jreader, err := operator.NewJsonReader(dr)
	if err != nil {
		if err == io.EOF {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("[]")) //If there are no datapoints, just return empty
			return nil
		}
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Errorln(err)
		return err
	}

	defer jreader.Close()
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(writer, jreader)
	if err != nil {
		logger.Errorln(err)
	}
	return nil
}

func GetStreamRange(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "StreamRange")
	q := request.URL.Query()

	i1s := q.Get("i1")
	i2s := q.Get("i2")

	//If either i1 or i2 are given, then it is a range by index
	if len(i1s) > 0 || len(i2s) > 0 {
		i1, err := strconv.ParseInt(i1s, 0, 64)
		if i1s != "" && err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			logger.Warningln(err)
			return ErrRangeArgs
		}

		i2, err := strconv.ParseInt(i2s, 0, 64)
		if i2s != "" && err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			logger.Warningln(err)
			return ErrRangeArgs
		} else if i2s == "" {
			i2 = 0
			i2s = "Inf"
		}

		logger.Debugf("irange [%s,%s)", i1s, i2s)

		dr, err := o.GetStreamIndexRange(streampath, i1, i2)

		return writeJSONResult(writer, dr, logger, err)
	}

	//It is not a range by index. See if it is a range by time
	t1s := q.Get("t1")
	t2s := q.Get("t2")
	if len(t1s) > 0 || len(t2s) > 0 {
		t1, err := strconv.ParseFloat(t1s, 64)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			logger.Warningln(err)
			return ErrRangeArgs
		}

		t2, err := strconv.ParseFloat(t2s, 64)
		if t2s != "" && err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			logger.Warningln(err)
			return ErrRangeArgs
		} else if t2s == "" {
			t2 = 0.
			t2s = "Inf"
		}

		lims := q.Get("limit")
		lim, err := strconv.ParseUint(lims, 0, 64)
		if lims != "" && err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			logger.Warningln(err)
			return ErrRangeArgs
		} else if lims == "" {
			lim = 0
			lims = "Inf"
		}

		logger.Debugf("trange [%s,%s) limit=%s", t1s, t2s, lims)
		dr, err := o.GetStreamTimeRange(streampath, t1, t2, int64(lim))
		if err != nil {
			writer.WriteHeader(http.StatusForbidden)
			logger.Warningln(err)
			return err
		}

		return writeJSONResult(writer, dr, logger, err)
	}

	//None of the limits were recognized. Rather than exploding, return bad request
	writer.WriteHeader(http.StatusBadRequest)
	logger.Warningln("Invalid range args")
	return ErrRangeArgs

}

//StreamTime2Index gets the time associated with the index
func StreamTime2Index(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := getStreamPath(request)
	logger = logger.WithField("op", "Time2Index")

	ts := request.URL.Query().Get("t")
	t, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warningln("invalid args")
		return ErrTime2IndexArgs
	}
	logger.Debugln("t=", ts)

	i, err := o.TimeToIndexStream(streampath, t)
	return JSONWriter(writer, i, logger, err)
}
