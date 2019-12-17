package assets

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// loadTestConfig takes the path to a conf file, or a folder, and loads all .conf files in order, merging them together.
func loadTestConfig(p string) (*Configuration, error) {
	s, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	if !s.IsDir() {
		return LoadConfigFile(p)
	}

	// It is a directory - list all files
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}
	confFiles := make([]string, 0, len(files))
	for i := range files {
		if strings.HasSuffix(files[i].Name(), ".conf") {
			confFiles = append(confFiles, files[i].Name())
		}

	}
	sort.Strings(confFiles)

	if len(confFiles) == 0 {
		return nil, errors.New("No configuration to load")
	}

	// Now load the configuration one by one:
	c := NewConfiguration()
	for i := range confFiles {
		tc, err := LoadConfigFile(path.Join(p, confFiles[i]))
		if err != nil {
			return nil, err
		}
		c = MergeConfig(c, tc)
	}

	return c, Validate(c)
}

func TestBadConfigurations(t *testing.T) {
	f, err := ioutil.ReadDir("tests/bad")
	require.NoError(t, err)
	for i := range f {
		filename := path.Join("tests/bad", f[i].Name())
		c, err := loadTestConfig(filename)
		if err == nil {
			err = Validate(c)
			require.Error(t, err, filename)
		}
	}
}

func TestGoodConfigurations(t *testing.T) {
	f, err := ioutil.ReadDir("tests/good")
	require.NoError(t, err)
	for i := range f {
		testFile := path.Join("tests/good", f[i].Name())
		c, err := loadTestConfig(testFile)
		require.NoError(t, err)
		require.NoError(t, Validate(c), testFile)
		c.URL = nil // The url is assumed to be nil, since it is automatically filled in on load
		b2, err := json.Marshal(c)
		require.NoError(t, err)
		resultFile := path.Join(testFile, "result.json")
		_, err = os.Stat(resultFile)
		if err == nil {
			b, err := ioutil.ReadFile(resultFile)
			require.NoError(t, err)
			require.JSONEq(t, string(b), string(b2), testFile+string(b2))
		}
	}
}

func TestBuiltinConfiguration(t *testing.T) {
	testreadconf := func(fname string) *Configuration {
		c, err := LoadConfigFile(fname)
		require.NoErrorf(t, err, "Could not read configuration file %s", fname)
		return c
	}

	c1 := testreadconf("../../assets/heedy.conf")
	c2 := testreadconf("../../assets/new/heedy.conf")

	MergeConfig(c1, c2)
}

func TestSchemaValidation(t *testing.T) {
	c, err := LoadConfigFile("../../assets/heedy.conf")
	require.NoError(t, err)

	v := map[string]interface{}{
		"schema": map[string]interface{}{},
		"actor":  false,
	}

	require.NoError(t, c.ValidateObjectMeta("stream", &v))

	v = map[string]interface{}{
		"actor": "hi",
	}
	require.Error(t, c.ValidateObjectMeta("stream", &v))

	v = map[string]interface{}{
		"ar": "hi",
	}
	require.Error(t, c.ValidateObjectMeta("stream", &v))
}
