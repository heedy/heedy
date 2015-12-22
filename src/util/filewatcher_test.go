/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package util

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestReloader struct {
	callnum int
}

func (tr *TestReloader) Reload() error {
	tr.callnum++
	return nil
}

func TestFileWatcher(t *testing.T) {
	ioutil.WriteFile("testfile.txt", []byte("test"), 0755)

	tr := TestReloader{}

	time.Sleep(50 * time.Millisecond)

	_, err := NewFileWatcher("testfile.txt", &tr)
	require.NoError(t, err)

	require.Equal(t, tr.callnum, 0)

	// Modification of file should call reload
	ioutil.WriteFile("testfile.txt", []byte("test"), 0755)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 1, tr.callnum)

	// Deleting a file should do nothing, but should call reload once it
	// is recreated
	require.NoError(t, os.Remove("testfile.txt"))
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 1, tr.callnum)

	ioutil.WriteFile("testfile.txt", []byte("test2"), 0755)
	time.Sleep(500 * time.Millisecond)
	require.Equal(t, 2, tr.callnum)

	// Now renaming a file, and then modifying it should do nothing
	os.Rename("testfile.txt", "testfile2.txt")
	ioutil.WriteFile("testfile2.txt", []byte("test3"), 0755)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 2, tr.callnum)

	// And lastly, renaming a file back to the original name should be caught
	os.Rename("testfile2.txt", "testfile.txt")
	time.Sleep(500 * time.Millisecond)
	require.Equal(t, 3, tr.callnum)
}
