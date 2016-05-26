/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"bytes"
	"io"

	"gopkg.in/vmihailenco/msgpack.v2"
)

//NewMsgPackDecoder returns our custom msgpack decoder
func NewMsgPackDecoder(r io.Reader) *msgpack.Decoder {
	dec := msgpack.NewDecoder(r)

	//Copied verbatim from msgpack decoder documentation
	dec.DecodeMapFunc = func(d *msgpack.Decoder) (interface{}, error) {
		n, err := d.DecodeMapLen()
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{}, n)
		for i := 0; i < n; i++ {
			mk, err := d.DecodeString()
			if err != nil {
				return nil, err
			}

			mv, err := d.DecodeInterface()
			if err != nil {
				return nil, err
			}

			m[mk] = mv
		}
		return m, nil
	}
	return dec
}

//MsgPackUnmarshal unmarshals msgpack with string keys in map
func MsgPackUnmarshal(b []byte, v ...interface{}) error {
	if len(v) == 1 && v[0] != nil {
		unmarshaler, ok := v[0].(msgpack.Unmarshaler)
		if ok {
			return unmarshaler.UnmarshalMsgpack(b)
		}
	}
	dec := NewMsgPackDecoder(bytes.NewReader(b))
	return dec.Decode(v...)
}
