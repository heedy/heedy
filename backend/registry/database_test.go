package registry

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	os.RemoveAll("./test.db")
	defer os.RemoveAll("./test.db")
	db, err := Create("./test.db")
	require.NoError(t, err)

	require.Equal(t, db.RegistryVersion, Version, "A newly created registry should have our version")

	tt := db.Updated

	require.NoError(t, db.Set("hello", "hi"))

	require.NoError(t, db.Close())

	db, err = Open("./test.db")
	require.NoError(t, err)

	require.Equal(t, db.Updated.Format(time.RFC3339), tt.Format(time.RFC3339))

	v, err := db.Get("hello")
	require.NoError(t, err)
	require.Equal(t, v, "hi")

	v, err = db.Get("noakey")
	require.NoError(t, err)
	require.Equal(t, v, "")

	db.Close()

}
