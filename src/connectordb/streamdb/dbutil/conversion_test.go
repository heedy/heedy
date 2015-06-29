package dbutil

import (
	"testing"
)

func TestGetConversion(t *testing.T) {
	// note that v0 should include all subsequent upgrades, so we should be fine here.

	_, err := getConversion(POSTGRES, "", false)
	if err != nil {
		t.Errorf("could not compile postgres conversion 0, no drop %v", err.Error())
	}

	_, err = getConversion(POSTGRES, "", true)
	if err != nil {
		t.Errorf("could not compile postgres conversion 0, drop %v", err.Error())
	}
}
