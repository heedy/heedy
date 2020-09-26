package timeseries

import (
	"errors"
	"time"

	"github.com/karrick/tparse"
)

func Unix(t time.Time) float64 {
	return float64(t.UnixNano()) * 1e-9
}

func ParseTimestamp(ts interface{}) (float64, error) {
	tss, ok := ts.(string)
	if ok {
		t, err := tparse.ParseNow(time.RFC3339, tss)
		return Unix(t), err
	}
	f, ok := ts.(float64)
	if ok {
		return f, nil
	}
	return 0, errors.New("Could not parse timestamp")
}
