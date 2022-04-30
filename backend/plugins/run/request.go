package run

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
)

type ResponseStreamer struct {
	reader *io.PipeReader
	writer *io.PipeWriter

	header http.Header

	Code int

	header_signal sync.Mutex
}

func (rs *ResponseStreamer) Header() http.Header {
	return rs.header
}

func (rs *ResponseStreamer) WriteHeader(code int) {
	if rs.Code == 0 {
		rs.Code = code

		rs.header_signal.Unlock()
	}
}

func (rs *ResponseStreamer) Write(p []byte) (n int, err error) {
	if rs.Code == 0 {
		rs.WriteHeader(http.StatusOK)
	}
	return rs.writer.Write(p)
}

// Runs the request in a goroutine, then waits until its header is written, and returns the resulting response code.
func (rs *ResponseStreamer) Serve(h http.Handler, r *http.Request) int {
	rs.header_signal.Lock()
	go func() {
		h.ServeHTTP(rs, r)
		if rs.Code == 0 {
			rs.WriteHeader(http.StatusOK)
		}
		rs.writer.Close()
	}()
	rs.header_signal.Lock() // The header signal is unlocked when the handler writes its header.
	return rs.Code
}

func (rs *ResponseStreamer) Read(p []byte) (n int, err error) {
	return rs.reader.Read(p)
}

func (rs *ResponseStreamer) Close() error {
	return rs.reader.Close()
}

func NewResponseStreamer() *ResponseStreamer {
	r, w := io.Pipe()
	return &ResponseStreamer{
		reader: r,
		writer: w,
		header: make(http.Header),
		Code:   0,
	}
}

func GetBodyReader(body any) (io.Reader, error) {
	var bodybuffer io.Reader
	if body != nil {
		switch body.(type) {
		case io.Reader:
			bodybuffer = body.(io.Reader)
		case []byte:
			bodybuffer = bytes.NewBuffer(body.([]byte))
		default:

			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodybuffer = bytes.NewBuffer(b)
		}
	}
	return bodybuffer, nil
}

func Request(h http.Handler, method, path string, body any, headers map[string]string) (*ResponseStreamer, error) {
	buf, err := GetBodyReader(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, path, buf)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}

	rs := NewResponseStreamer()

	rs.Serve(h, req)

	// If the return code is a 4xx or 5xx, return the error
	if rs.Code >= 400 {
		var er rest.ErrorResponse
		//Limit requests to the limit given in configuration
		data, err := ioutil.ReadAll(io.LimitReader(rs, *assets.Config().RequestBodyByteLimit))
		if err != nil {
			rs.Close()
			return nil, err
		}
		err = json.Unmarshal(data, &er)
		if err != nil {
			return nil, err
		}
		return nil, &er
	}

	// Otherwise return the response streamer, which is really an io.ReadCloser
	return rs, nil
}
