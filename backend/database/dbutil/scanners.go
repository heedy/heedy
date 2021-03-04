package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	t := fmt.Sprintf("\"%s\"", d.String())
	return []byte(t), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	*d = Date(t)

	return err
}

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

type JSONArray struct {
	Elements []interface{}
}

func (ja *JSONArray) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &ja.Elements)
	case string:
		return json.Unmarshal([]byte(v), &ja.Elements)
	default:
		return fmt.Errorf("Can't scan json array array, unsupported type: %T", v)
	}
}

func (ja *JSONArray) Value() (driver.Value, error) {
	return ja.MarshalJSON()
}

func (ja *JSONArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(ja.Elements)
}

func (ja *JSONArray) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ja.Elements)
}

type StringArray struct {
	Strings []string
	sMap    map[string]bool
}

func (s *StringArray) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &s.Strings)
	case string:
		return json.Unmarshal([]byte(v), &s.Strings)
	default:
		return fmt.Errorf("Can't scan string array, unsupported type: %T", v)
	}
}

func (s *StringArray) Load(total string) {
	s.Strings = strings.Fields(total)
	s.sMap = nil // Clear the old map
	s.Deduplicate()
}

func (s *StringArray) String() string {
	return strings.Join(s.Strings, " ")
}

func (s *StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StringArray) UnmarshalJSON(b []byte) error {
	var total string
	err := json.Unmarshal(b, &total)
	if err == nil {
		s.Load(total)
	}
	return err
}

func (s *StringArray) Value() (driver.Value, error) {
	return json.Marshal(s.Strings)
}

func (s *StringArray) LoadMap() {
	if s.sMap == nil {
		smap := make(map[string]bool)
		for _, v := range s.Strings {
			smap[v] = true
		}
		s.sMap = smap
	}
}
func (s *StringArray) Deduplicate() {
	s.LoadMap()
	s.Strings = make([]string, 0, len(s.sMap))
	for k := range s.sMap {
		s.Strings = append(s.Strings, k)
	}
}

func (s *StringArray) Contains(v string) bool {
	s.LoadMap()
	_, ok := s.sMap[v]
	return ok
}

func (s *StringArray) HasSubset(s2 []string) bool {
	s.LoadMap()
	for _, v := range s2 {
		if !s.Contains(v) {
			return false
		}
	}
	return true
}

// JSONObject represents a json column in the table. To handle it correctly, we need to manually scan it
// and output the relevant values
type JSONObject map[string]interface{}

func (s *JSONObject) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s)
		return nil
	case string:
		json.Unmarshal([]byte(v), &s)
		return nil
	default:
		return fmt.Errorf("Can't unmarshal json object, unsupported type: %T", v)
	}
}
func (s *JSONObject) Value() (driver.Value, error) {
	return json.Marshal(s)
}
