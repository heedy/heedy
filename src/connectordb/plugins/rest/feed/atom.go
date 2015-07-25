package feed

//Code based upon golang.org/x/tools/blog/atom

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"connectordb/plugins/rest/restcore"

	log "github.com/Sirupsen/logrus"
)

func AtomTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

type Person struct {
	Name string `xml:"name"`
}

type Text struct {
	Type string `xml:"type,attr,omitempty"`
	Body string `xml:",chardata"`
}

type Link struct {
	Href string `xml:"href,attr"`
}

type Feed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Link    Link     `xml:"link"`
	Updated string   `xml:"updated"`
	Author  *Person  `xml:"author"`
	Entry   []*Entry `xml:"entry"`
}

type Entry struct {
	Title   string  `xml:"title"`
	ID      string  `xml:"id"`
	Link    Link    `xml:"link"`
	Updated string  `xml:"updated"`
	Author  *Person `xml:"author"`
	Summary *Text   `xml:"summary"`
}

func getAtomEntry(dpindex int64, dp datastream.Datapoint, streamname string) *Entry {
	return nil
}

//GetAtom gets an Atom feed of the given stream.
func GetAtom(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	usrname, devname, _, streampath := restcore.GetStreamPath(request)
	s, dr, err := getFeedData(o, writer, request, logger)
	if err != nil {
		return err
	}

	f := Feed{
		Title:   streampath,
		ID:      strconv.FormatInt(s.StreamId, 10),
		Updated: AtomTime(time.Now()),
		Author:  &Person{usrname},
		Link:    Link{Href: "https://connectordb.com/api/v1/d/" + streampath}, //I dislike links. Especially hard-coded ones
		Entry:   make([]*Entry, 0, EntryLimit),
	}
	i := dr.Index()
	for dp, err := dr.Next(); err == nil && dp != nil; dp, err = dr.Next() {
		v, err := json.Marshal(dp.Data)
		if err != nil {
			restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
			return err
		}
		authr := dp.Sender
		if authr == "" {
			authr = usrname + "/" + devname
		}

		f.Entry = append(f.Entry, &Entry{
			Updated: AtomTime(time.Unix(0, int64(dp.Timestamp*1e9))),
			Title:   "",
			Link:    Link{Href: "https://connectordb.com/api/v1/d/" + streampath + "/data?i1=" + strconv.FormatInt(i, 10) + "&i2=" + strconv.FormatInt(i+1, 10)},
			ID:      strconv.FormatInt(s.StreamId, 10) + ":" + strconv.FormatInt(i, 10),
			Author:  &Person{authr},
			Summary: &Text{Body: string(v)},
		})
		i = dr.Index()

	}

	result, err := xml.Marshal(f)
	if err != nil {
		restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
		return err
	}

	writer.Header().Set("Content-Length", strconv.Itoa(len(result)))
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write(result)
	return nil
}
