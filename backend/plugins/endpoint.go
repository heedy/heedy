package plugins

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

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
func WaitForEndpoint(method string, host string, e *Exec) error {
	// The endpoint is not available, so let's keep checking it
	d := 30 * time.Second
	sleepDuration := 100 * time.Millisecond
	for i := time.Duration(0); i < d; i += sleepDuration {
		c, err := net.Dial(method, host)
		if err == nil {
			c.Close()
			return nil
		}
		if err = e.HadError(); err != nil {
			return err
		}
		time.Sleep(sleepDuration)
	}
	return fmt.Errorf("Could not connect to %s using %s socket", host, method)
}
