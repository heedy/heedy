package plugin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type unixDialer struct {
	Location string
	net.Dialer
}

// overriding net.Dialer.Dial to force unix socket connection
func (d *unixDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dialer.DialContext(ctx, "unix", d.Location)
}

// NewReverseProxy generates a reverse proxy from a given uri, automatically handling unix uris
func NewReverseProxy(datadir, uri string) (*httputil.ReverseProxy, error) {

	if !strings.HasPrefix(uri, "unix://") {
		parsedURL, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(parsedURL), nil
	}

	// Otherwise, we set up a unix domain socket.

	splitopath := strings.SplitAfterN(uri[7:], ".sock", 2)
	host := splitopath[0]
	if !strings.HasSuffix(host, ".sock") {
		return nil, fmt.Errorf("A unix socket must have its file end with .sock ('%s')", uri)
	}
	path := splitopath[1]
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("The url after .sock must start with / ('%s')", uri)
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

	if !filepath.IsAbs(host) {
		host = filepath.Join(datadir, host)
	}

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
