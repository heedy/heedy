package run

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/heedy/heedy/api/golang/rest"
)

type unixDialer struct {
	Location string
	net.Dialer
}

// overriding net.Dialer.Dial to force unix socket app
func (d *unixDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dialer.DialContext(ctx, "unix", d.Location)
}

// NewReverseProxy generates a reverse proxy from a given uri, automatically handling unix sockets,
// and builtin handlers
func NewReverseProxy(datadir, uri string) (http.Handler, error) {

	gatewayError := func(w http.ResponseWriter, r *http.Request, err error) {
		rest.WriteJSONError(w, r, http.StatusBadGateway, fmt.Errorf("plugin_error: %s", err.Error()))
	}

	if !strings.HasPrefix(uri, "unix:") {
		parsedURL, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		p := httputil.NewSingleHostReverseProxy(parsedURL)
		p.ErrorHandler = gatewayError
		return p, nil
	}

	// Otherwise, we set up a unix domain socket.
	host, path, err := ParseUnixSock(datadir, uri)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Host:   host,
		Scheme: "http",
		Path:   "/",
	}
	parsedURL, err := u.Parse(path)
	if err != nil {
		return nil, err
	}

	p := httputil.NewSingleHostReverseProxy(parsedURL)
	p.ErrorHandler = gatewayError

	p.Transport = &http.Transport{
		DialContext: (&unixDialer{
			Dialer: net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			},
			Location: host,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return p, nil

}
