package run

/*
import (
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/stretchr/testify/require"
)

func TestReverseProxy(t *testing.T) {
	a, err := assets.Open("./tests/unix_revproxy", nil)
	require.NoError(t, err)

	os.MkdirAll(path.Join(a.FolderPath, "data"), 0775)

	p, err := NewReverseProxy(a.DataDir(), "unix:test/server.sock/test")
	require.NoError(t, err)

	m := NewExecManager(a)
	require.NoError(t, m.Start())
	time.Sleep(1 * time.Second)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/me", nil)

	p.ServeHTTP(rec, req)

	m.Stop()

	require.Equal(t, rec.Code, 200)

	require.Equal(t, "Hello World!", rec.Body.String())

}
*/
