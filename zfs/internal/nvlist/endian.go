package nvlist

import (
	"encoding/binary"
	"fmt"
)

type Endian int8

const (
	LittleEndian = Endian(iota) // 0
	BigEndian                   // 1
)

func (e Endian) ByteOrder() binary.ByteOrder {
	return nil
}

func (e Endian) String() string {
	switch e {
	case BigEndian:
		return "BigEndian"
	case LittleEndian:
		return "LittleEndian"
	}
	return fmt.Sprintf("*ERROR-INVALID-ENDIAN-VALUE-%02x*", int8(e))
}
