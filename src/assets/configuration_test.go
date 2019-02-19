package assets

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfiguration(t *testing.T) {
	fmt.Printf("\nBUILTIN CONFIGURATIONS ---------------------------\n\n")
	testreadconf := func(fname string) {
		c, err := LoadConfigFile("../../assets/connectordb.conf")
		require.NoErrorf(t, err, "Could not read configuration file %s", fname)

		b, err := json.MarshalIndent(c, "", "  ")
		require.NoErrorf(t, err, "Couldn't marshal %s", fname)
		fmt.Printf("%s\n%s\n\n", fname, string(b))
	}

	testreadconf("../../assets/connectordb.conf")
	testreadconf("../../assets/newdb/connectordb.conf")

	fmt.Printf("------------------------------------------------------\n\n")
}
