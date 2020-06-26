package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/sirupsen/logrus"
)

// Request runs the given http handler, if body is byte array, sends that, otherwise marshals as json, and optionally unmarshals the result
func Request(h http.Handler, method, path string, body interface{}, headers map[string]string) (*bytes.Buffer, error) {
	var bodybuffer io.Reader
	if body != nil {
		b, ok := body.([]byte)
		if !ok {
			var err error
			b, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}

		bodybuffer = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, path, bodybuffer)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code >= 400 {
		var er rest.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &er)
		if err != nil {
			return nil, err
		}
		return nil, &er
	}
	return rec.Body, nil
}

func Route(m *chi.Mux, route string, h http.Handler) error {
	ss := strings.Fields(route)
	if len(ss) == 0 || len(ss) > 2 {
		return errors.New("Invalid route")
	}
	if len(ss) == 1 {
		m.Handle(ss[0], h)
		return nil
	}
	switch ss[0] {
	case "GET":
		m.Get(ss[1], h.ServeHTTP)
	case "POST":
		m.Post(ss[1], h.ServeHTTP)
	case "PUT":
		m.Put(ss[1], h.ServeHTTP)
	case "DELETE":
		m.Delete(ss[1], h.ServeHTTP)
	case "PATCH":
		m.Patch(ss[1], h.ServeHTTP)
	default:
		return errors.New("Unrecognized verb")
	}
	return nil
}

func GetPlugin(plugin, uri string) (string, string, string) {
	if !strings.HasPrefix(uri, "run://") {
		return "", "", ""
	}

	// The uri starts with run://, which means that it is referring to a runner.
	splitstring := strings.SplitN(uri[6:len(uri)], "/", 2)
	pluginv := strings.SplitN(splitstring[0], ":", 2)
	pname := pluginv[0]
	if len(pluginv) > 1 {
		pname = pluginv[1]
		plugin = pluginv[0]
	}

	hpath := "/"
	if len(splitstring) > 1 {
		hpath = hpath + splitstring[1]
	}
	return plugin, pname, hpath
}

// Extracts the unix socket file and request path
func ParseUnixSock(datadir string, uri string) (sockfile string, requestPath string, err error) {
	if !strings.HasPrefix(uri, "unix://") {
		err = errors.New("Not a unix socket")
		return
	}
	// Otherwise, we set up a unix domain socket.

	splitopath := strings.SplitAfterN(uri[7:], ".sock", 2)
	sockfile = splitopath[0]
	if !strings.HasSuffix(sockfile, ".sock") {
		err = fmt.Errorf("A unix socket must have its file end with .sock ('%s')", uri)
		return
	}
	if !filepath.IsAbs(sockfile) {
		sockfile = filepath.Join(datadir, sockfile)
	}

	requestPath = splitopath[1]
	if requestPath == "" {
		requestPath = "/"
	}
	if !strings.HasPrefix(requestPath, "/") {
		err = fmt.Errorf("The url after .sock must start with / ('%s')", uri)
	}
	return
}

// GetEndpoint parses the given URI and returns an endpoint
func GetEndpoint(datadir string, uri string) (method string, host string, err error) {
	if strings.HasPrefix(uri, "unix://") {
		method = "unix"
		host, _, err = ParseUnixSock(datadir, uri)
	} else {
		var u *url.URL
		u, err = url.Parse(uri)
		if err != nil {
			return
		}
		host = u.Host
		method = "tcp"
	}
	return
}

// WaitForEndpoint waits for the given endpoint
func WaitForEndpoint(method string, host string, e *Cmd) error {
	logrus.Debugf("Waiting for %s://%s", method, host)
	// The endpoint is not available, so let's keep checking it
	d := 30 * time.Second
	sleepDuration := 100 * time.Millisecond
	for i := time.Duration(0); i < d; i += sleepDuration {
		c, err := net.Dial(method, host)
		if err == nil {
			c.Close()
			logrus.Debugf("endpoint open %s://%s", method, host)
			return nil
		}
		if e.Done() {
			return errors.New("Process exited")
		}
		time.Sleep(sleepDuration)
	}
	return fmt.Errorf("Could not connect to %s using %s socket", host, method)
}

// WaitForAPI is like WaitForEndpoint, but it doesn't have a cmd.
func WaitForAPI(method string, host string, timeout time.Duration) error {
	logrus.Debugf("Waiting for %s://%s", method, host)
	sleepDuration := 100 * time.Millisecond
	for i := time.Duration(0); i < timeout; i += sleepDuration {
		c, err := net.Dial(method, host)
		if err == nil {
			c.Close()
			logrus.Debugf("endpoint open %s://%s", method, host)
			return nil
		}
		time.Sleep(sleepDuration)
	}
	return fmt.Errorf("Could not connect to %s using %s socket", host, method)
}
