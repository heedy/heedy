package server

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

const cacheControlStatic = "max-age=0,stale-while-revalidate=604800"

func setCacheControl(w http.ResponseWriter, r *http.Request, noCache bool) {
	if noCache {
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		w.Header().Set("Cache-Control", cacheControlStatic)
	}
}

// Based on
// https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/net/http/fs.go
// https://github.com/lpar/gzipped/blob/v1.1.0/fileserver.go

type staticHandler struct {
	fs      http.FileSystem
	noCache bool
}

// NewStaticHandler tries to serve the original files, but if they are not found,
// it will serve a compressed version of the file. Usually this is reversed, but in heedy,
// the build process actually *removes* the original files, so when the compressed files exist,
// the originals are not present. It also sets up the correct caching headers for static files
func NewStaticHandler(root http.FileSystem, nocache bool) http.Handler {
	return &staticHandler{fs: root, noCache: nocache}
}

// https://github.com/lpar/gzipped/blob/v1.1.0/fileserver.go#L45
func acceptable(r *http.Request, encoding string) bool {
	for _, aspec := range header.ParseAccept(r.Header, "Accept-Encoding") {
		if aspec.Value == encoding && aspec.Q == 0.0 {
			return false
		}
		if (aspec.Value == encoding || aspec.Value == "*") && aspec.Q > 0.0 {
			return true
		}
	}
	return false
}

func (pch *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
		r.URL.Path = name
	}
	name = path.Clean(name)
	if strings.HasSuffix(name, "/") {
		http.NotFound(w, r)
		return
	}

	f, err := pch.fs.Open(name)
	isgzip := false
	if errors.Is(err, fs.ErrNotExist) {
		// We don't accept range queries over the gzipped files.
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
			return
		}
		f, err = pch.fs.Open(name + ".gz")
		if err == nil {
			// The gzipped file was found. Now check if the client can actually accept it...
			if !acceptable(r, "gzip") {
				http.Error(w, "406 Not Acceptable", http.StatusNotAcceptable)
				return
			}

			isgzip = true
		}
	}
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// Is it a folder?
	if d.IsDir() {
		// We don't allow directory listing.
		http.NotFound(w, r)
		return
	}

	if isgzip {
		// ServeContent doesn't actually set the content length if there is a content-encoding header.
		w.Header().Set("Content-Length", strconv.FormatInt(d.Size(), 10))
		w.Header().Set("Content-Encoding", "gzip")
	}

	// Set caching headers
	if pch.noCache {
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		w.Header().Set("Cache-Control", cacheControlStatic)
	}

	http.ServeContent(w, r, name, d.ModTime(), f)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if errors.Is(err, fs.ErrNotExist) {
		return "404 page not found", http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}
