package assets

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfiguration(t *testing.T) {
	fmt.Printf("\nBUILTIN CONFIGURATIONS ---------------------------\n\n")
	testreadconf := func(fname string) *Configuration {
		c, err := LoadConfigFile(fname)
		require.NoErrorf(t, err, "Could not read configuration file %s", fname)

		b, err := json.MarshalIndent(c, "", "  ")
		require.NoErrorf(t, err, "Couldn't marshal %s", fname)
		fmt.Printf("%s\n%s\n\n", fname, string(b))
		return c
	}

	c1 := testreadconf("../../assets/heedy.conf")
	c2 := testreadconf("../../assets/new/heedy.conf")

	fmt.Printf("MERGED------------------------------------------------\n\n")

	c3 := MergeConfig(c1, c2)
	b, err := json.MarshalIndent(c3, "", "  ")
	require.NoErrorf(t, err, "Couldn't marshal merged config")
	fmt.Printf("%s\n\n", string(b))

	fmt.Printf("------------------------------------------------------\n\n")
}
