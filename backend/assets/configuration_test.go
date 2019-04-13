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
		c, err := LoadConfigFile(fname)
		require.NoErrorf(t, err, "Could not read configuration file %s", fname)

		b, err := json.MarshalIndent(c, "", "  ")
		require.NoErrorf(t, err, "Couldn't marshal %s", fname)
		fmt.Printf("%s\n%s\n\n", fname, string(b))
	}

	testreadconf("../../assets/heedy.conf")
	testreadconf("../../assets/new/heedy.conf")

	fmt.Printf("------------------------------------------------------\n\n")
}
