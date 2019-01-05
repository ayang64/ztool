package nvlist

import (
	"encoding/binary"
	"io"
	"log"
)

type Header struct {
	Encoding  int8
	Endian    int8
	Reserved1 int8
	Reserved2 int8
}

type List struct {
	Version int32
	Flags   uint32
	Private uint64
	Flag    uint32
}

const (
	HeaderMagic = 0x6c
)

// ensure the header is valid
func (h Header) Verify() error {
	return nil
}

type Type uint32

// nvlist data types
const (
	Unknown = Type(iota)
	Boolean
	Byte
	Int16
	Uint16
	Int32
	Uint32
	String
	ByteArray
	Uint16Array
	Int32Array
	Int64Array
	StringArray
	HRTime
	NVList
	NVListArray
	BooleanValue
	Int8
	Uint8
	BooleanArray
	Int8Array
	Uint8Array
	Double
)

func getRecord(r io.Reader) (string, []byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", nil, err
	}

	log.Printf("length: %d", length)

	var keylen uint32
	if err := binary.Read(r, binary.BigEndian, &keylen); err != nil {
		return "", nil, err
	}

	log.Printf("key length: %d", keylen)

	key := make([]byte, keylen)
	if err := binary.Read(r, binary.BigEndian, &key); err != nil {
		return "", nil, err
	}

	log.Printf("key: %q", string(key))

	datalen := length - 4 - keylen
	log.Printf("datalen: %d", datalen)

	data := make([]byte, 0, 0)
	if err := binary.Read(r, binary.BigEndian, &data); err != nil {
		return "", nil, err
	}

	return string(key), nil, nil
}

func DumpNvlist(r io.Reader) {
	log.Printf("DUMPING NVLIST")
	for {
		key, data, err := getRecord(r)

		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		data = data
		key = key
	}
}
