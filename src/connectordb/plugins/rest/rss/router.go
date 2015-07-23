package rss

import (
	"connectordb/plugins/rest/restcore"
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

//GetRSS gets an RSS feed of the given stream. It returns the smaller of 2 days or 500 datapoints
func GetRSS(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	_, _, _, streampath := restcore.GetStreamPath(request)

	tnow := time.Now()
	t2day := tnow.Add(-2 * 24 * time.Hour)

	s, err := o.ReadStream(streampath)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusForbidden, err, false)
		return err
	}

	dr, err := o.GetStreamTimeRange(streampath, float64(t2day.Unix()), float64(tnow.Unix()), int64(500))
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
		return err
	}

	f := &feeds.Feed{
		Title:       s.Name,
		Created:     time.Now(),
		Description: s.Nickname,
		Link:        &feeds.Link{Href: "https://connectordb.com/"},
	}

	f.Items = make([]*feeds.Item, 0, 500)

	for dp, err := dr.Next(); err == nil && dp != nil; dp, err = dr.Next() {
		v, err := json.Marshal(dp.Data)
		if err != nil {
			restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
			return err
		}
		f.Items = append(f.Items, &feeds.Item{
			Created: time.Unix(0, int64(dp.Timestamp*1e9)),
			Link:    &feeds.Link{Href: ""},
			Title:   string(v),
		})
	}

	val, err := f.ToAtom()
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
		return err
	}

	writer.Header().Set("Content-Length", strconv.Itoa(len([]byte(val))))
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(val))

	return nil
}

//Router returns a fully formed Gorilla router given an optional prefix
func Router(db *streamdb.Database, prefix *mux.Router) *mux.Router {
	if prefix == nil {
		prefix = mux.NewRouter()
	}

	//Allow for the application to match /path and /path/ to the same place.
	prefix.StrictSlash(true)

	prefix.HandleFunc("/{user}/{device}/{stream}", restcore.Authenticator(GetRSS, db)).Methods("GET")

	return prefix
}
