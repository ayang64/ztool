package nvlist

import (
	"encoding/binary"
	"io"
)

type Encoder struct {
	bo binary.ByteOrder
	w  io.Writer
}

func NewEncoder(w io.Writer, bo binary.ByteOrder) *Encoder {
	return &Encoder{
		bo: bo,
		w:  w,
	}
}

func (e *Encoder) EncodeString(w io.Writer, s string) error {
	l := int32(len(s))
	binary.Write(e.w, e.bo, l)
	binary.Write(e.w, e.bo, s)
	return nil
}

func (e *Encoder) Encode(w io.Writer, m map[string]interface{}) {
	/*
		for k := range m {
			switch v := m[k].(type) {
			case uint64:
			case string:
			}
		}
	*/
}
