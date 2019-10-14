package updater

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestUpdateHeedy(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	require.NoError(t, os.MkdirAll("./tester/updates", 0775))
	require.NoError(t, os.MkdirAll("./tester/updates.failed", 0775))
	require.NoError(t, os.MkdirAll("./tester/backup", 0775))
	defer os.RemoveAll("./tester")

	ioutil.WriteFile("./tester/heedy", []byte("current"), 0664)
	ioutil.WriteFile("./tester/updates/heedy", []byte("next"), 0664)
	ioutil.WriteFile("./tester/updates/heedy.sig", []byte("sig"), 0664)

	require.NoError(t, UpdateHeedy("tester", "tester/updates", "tester/backup"))

	b, err := ioutil.ReadFile("./tester/heedy")
	require.NoError(t, err)
	require.Equal(t, string(b), "next")

	b, err = ioutil.ReadFile("./tester/backup/heedy")
	require.NoError(t, err)
	require.Equal(t, string(b), "current")

	// Now revert the replace
	require.NoError(t, RevertHeedy("tester", "tester/backup", "tester/updates.failed"))

	b, err = ioutil.ReadFile("./tester/heedy")
	require.NoError(t, err)
	require.Equal(t, string(b), "current")

	b, err = ioutil.ReadFile("./tester/updates.failed/heedy")
	require.NoError(t, err)
	require.Equal(t, string(b), "next")
}

func TestUpdatePlugins(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	require.NoError(t, os.MkdirAll("./tester/plugins/testy", 0775))
	require.NoError(t, os.MkdirAll("./tester/updates/plugins/testy", 0775))
	require.NoError(t, os.MkdirAll("./tester/updates.failed", 0775))
	require.NoError(t, os.MkdirAll("./tester/backup", 0775))
	defer os.RemoveAll("./tester")

	ioutil.WriteFile("./tester/plugins/testy/heedy.conf", []byte("current"), 0664)
	ioutil.WriteFile("./tester/updates/plugins/testy/heedy.conf", []byte("next"), 0664)

	require.NoError(t, UpdatePlugins("tester", "tester/updates", "tester/backup"))

	b, err := ioutil.ReadFile("./tester/plugins/testy/heedy.conf")
	require.NoError(t, err)
	require.Equal(t, string(b), "next")

	b, err = ioutil.ReadFile("./tester/backup/plugins/testy/heedy.conf")
	require.NoError(t, err)
	require.Equal(t, string(b), "current")

	require.NoError(t, RevertPlugins("tester", "tester/backup", "tester/updates.failed"))

	b, err = ioutil.ReadFile("./tester/plugins/testy/heedy.conf")
	require.NoError(t, err)
	require.Equal(t, string(b), "current")

	b, err = ioutil.ReadFile("./tester/updates.failed/plugins/testy/heedy.conf")
	require.NoError(t, err)
	require.Equal(t, string(b), "next")

}

func TestBackupData(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	require.NoError(t, os.MkdirAll("./tester/data/td", 0775))
	require.NoError(t, os.MkdirAll("./tester/updates.failed", 0775))
	require.NoError(t, os.MkdirAll("./tester/backup", 0775))
	defer os.RemoveAll("./tester")

	ioutil.WriteFile("./tester/data/data.db", []byte("data"), 0664)
	ioutil.WriteFile("./tester/data/td/blah", []byte("blah"), 0664)

	require.NoError(t, BackupData("tester", "tester/updates", "tester/backup"))

	ioutil.WriteFile("./tester/data/data.db", []byte("bad_data"), 0664)
	ioutil.WriteFile("./tester/data/td/blah", []byte("bad_blah"), 0664)

	b, err := ioutil.ReadFile("./tester/data/data.db")
	require.NoError(t, err)
	require.Equal(t, string(b), "bad_data")

	b, err = ioutil.ReadFile("./tester/data/td/blah")
	require.NoError(t, err)
	require.Equal(t, string(b), "bad_blah")

	require.NoError(t, RevertData("tester", "tester/backup", "tester/updates.failed"))

	b, err = ioutil.ReadFile("./tester/data/data.db")
	require.NoError(t, err)
	require.Equal(t, string(b), "data")

	b, err = ioutil.ReadFile("./tester/data/td/blah")
	require.NoError(t, err)
	require.Equal(t, string(b), "blah")

}
