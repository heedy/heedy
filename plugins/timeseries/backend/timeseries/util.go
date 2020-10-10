package timeseries

import (
	"errors"
	"time"

	"github.com/karrick/tparse"
	jwriter "github.com/mailru/easyjson/jwriter"
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

func jsonInterfaceMarshaller(out *jwriter.Writer, in interface{}) {
	if in == nil {
		out.RawString("null")
		return
	}
	switch v := in.(type) {
	case string:
		out.String(v)
	case float64:
		out.Float64(v)
	case float32:
		out.Float32(v)
	case int:
		out.Int(v)
	case int64:
		out.Int64(v)
	case bool:
		out.Bool(v)
	case map[string]interface{}:
		curb := byte('{')
		for k, vv := range v {
			out.RawByte(curb)
			out.String(k)
			out.RawByte(':')
			jsonInterfaceMarshaller(out, vv)
			curb = ','
		}
		out.RawByte('}')
	case []interface{}:
		curb := byte('[')
		for _, vv := range v {
			out.RawByte(curb)
			jsonInterfaceMarshaller(out, vv)
			curb = ','
		}
		out.RawByte(']')
	default:
		out.Error = errors.New("Unknown data type when encoding point data to json")
	}
}
