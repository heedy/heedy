package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/stretchr/testify/require"
)

func TestExec(t *testing.T) {
	a, err := assets.Open("./tests/exec", nil)
	require.NoError(t, err)

	os.MkdirAll(path.Join(a.FolderPath, "data"), 0775)

	m := NewExecManager(a)
	require.NoError(t, m.Start())
	time.Sleep(time.Second)
	require.NoError(t, m.Stop())

	v, err := ioutil.ReadFile(path.Join(a.FolderPath, "data", "test.txt"))
	require.NoError(t, err)
	exc := &Exec{}
	require.NoError(t, json.Unmarshal(v, exc))

	for k := range m.Processes {
		require.Equal(t, k, exc.APIKey)
	}

}
