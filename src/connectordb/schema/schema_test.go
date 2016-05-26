/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package schema

import (
	"testing"
)

func TestSchema(t *testing.T) {
	s_string, err := NewSchema(`{"type": "string"}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}
	s_float, err := NewSchema(`{"type": "number"}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}
	s_obj, err := NewSchema(`{"type": "object", "properties": {"lat": {"type": "number"},"msg": {"type": "string"}}}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}

	v_string := "Hello"
	v_float := 3.14
	v_obj := map[string]interface{}{"lat": 88.32, "msg": "hi"}
	v_bobj := map[string]interface{}{"lat": "88.32", "msg": "hi"}
	duk_obj := map[string]interface{}{"lat": 88.32, "msg": "hi", "testing": 123}

	if !s_string.IsValid(v_string) || !s_float.IsValid(v_float) || !s_obj.IsValid(v_obj) {
		t.Errorf("Validation failed")
		return
	}
	if s_obj.IsValid(v_bobj) {
		t.Errorf("Validation wrong")
		return
	}

	if !s_obj.IsValid(duk_obj) {
		t.Errorf("Validation for object %v with schema %v failed", duk_obj, s_obj)
		return
	}

	var x interface{}
	var z interface{}

	val, err := s_string.Marshal(v_string)
	if err != nil {
		t.Errorf("Marshal failed")
		return
	}

	if err = s_string.Unmarshal(val, &x); err != nil {
		t.Errorf("unmarshal failed")
		return
	}
	if v, ok := x.(string); !ok || v != v_string {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}

	val, err = s_float.Marshal(v_float)
	if err != nil {
		t.Errorf("Marshal failed %v", err)
		return
	}

	if err = s_float.Unmarshal(val, &z); err != nil {
		t.Errorf("unmarshal failed %v", err)
		return
	}
	if v, ok := z.(float64); !ok || v != v_float {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}

	val, err = s_obj.Marshal(v_obj)
	if err != nil {
		t.Errorf("Marshal failed")
		return
	}

	//msgpack is weird in that it sees the values of previous interfaces - so we need to create a new uninitialized interface here
	var xo interface{}

	if err = s_obj.Unmarshal(val, &xo); err != nil {
		t.Errorf("unmarshal failed")
		return
	}
	if v, ok := xo.(map[string]interface{}); !ok || v["lat"].(float64) != 88.32 || v["msg"].(string) != "hi" {
		t.Errorf("Crap: %v, %v", ok, v)
		return
	}
}
