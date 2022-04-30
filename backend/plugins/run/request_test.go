package run

import (
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	fnc := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello World!"))
		resp.Write([]byte("World2"))
	})

	resp, err := Request(fnc, "GET", "/lol", nil, nil)
	require.NoError(t, err)
	b, err := ioutil.ReadAll(resp)
	require.NoError(t, err)
	require.Equal(t, "Hello World!World2", string(b))
}

func TestRequestInterrupted(t *testing.T) {
	lock := sync.Mutex{}
	lock.Lock()
	var werr error
	var wint int
	fnc := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		wint, werr = resp.Write([]byte("Hello World!"))
		lock.Unlock()
	})
	resp, err := Request(fnc, "GET", "/lol", nil, nil)
	require.NoError(t, err)
	buf := make([]byte, 5)
	n, err := resp.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, "Hello", string(buf))
	resp.Close()
	lock.Lock()
	require.Equal(t, 5, wint)
	require.Error(t, werr)

}
