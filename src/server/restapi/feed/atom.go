package feed

//Code based upon golang.org/x/tools/blog/atom

import (
	"connectordb/datastream"
	"connectordb/operator"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"time"

	"server/restapi/restcore"
	"server/webcore"

	log "github.com/Sirupsen/logrus"
	"github.com/connectordb/duck"
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
	Rel  string `xml:"rel,attr,omitempty"`
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
	Content *Text   `xml:"content"`
}

func getAtomEntry(dpindex int64, dp datastream.Datapoint, streamname string) *Entry {
	return nil
}

//GetAtom gets an Atom feed of the given stream.
func GetAtom(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	usrname, devname, _, streampath := restcore.GetStreamPath(request)
	_, dr, err := getFeedData(o, writer, request, logger)
	if err != nil {
		return webcore.INFO, ""
	}
	sdr, ok := dr.(datastream.ExtendedDataRange)
	if !ok {
		return restcore.WriteError(writer, logger, http.StatusInternalServerError, errors.New("Internal server error: Unable to convert DataRange to ExtendedDataRange"), true)
	}

	streamuri := webcore.SiteName + "/api/v1/feed/" + streampath + ".atom"
	f := Feed{
		Title:   streampath,
		ID:      streamuri,
		Updated: AtomTime(time.Now()),
		Author:  &Person{usrname},
		Link:    Link{Href: streamuri, Rel: "self"}, //I dislike links. Especially hard-coded ones
		Entry:   make([]*Entry, 0, EntryLimit),
	}
	i := sdr.Index()
	for dp, err := dr.Next(); err == nil && dp != nil; dp, err = dr.Next() {
		v := duck.JSONString(dp.Data)
		if v == "" {
			return restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
		}
		authr := dp.Sender
		if authr == "" {
			authr = usrname + "/" + devname
		}

		feeduri := webcore.SiteName + "/api/v1/crud/" + streampath + "/data?i1=" + strconv.FormatInt(i, 10) + "&i2=" + strconv.FormatInt(i+1, 10)

		f.Entry = append(f.Entry, &Entry{
			Updated: AtomTime(time.Unix(0, int64(dp.Timestamp*1e9))),
			Title:   "Datapoint " + strconv.FormatInt(i+1, 10),
			Link:    Link{Href: feeduri},
			ID:      feeduri,
			Author:  &Person{authr},
			Content: &Text{Body: v},
		})
		i = sdr.Index()
	}

	result, err := xml.Marshal(f)
	if err != nil {
		return restcore.WriteError(writer, logger, http.StatusInternalServerError, err, true)
	}
	xmlheader := []byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")

	writer.Header().Set("Content-Length", strconv.Itoa(len(result)+len(xmlheader)))
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write(xmlheader)
	writer.Write(result)
	return webcore.DEBUG, ""
}
