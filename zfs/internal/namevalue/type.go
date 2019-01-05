package nvlist

import (
	"encoding/binary"
	"io"
	"log"
)

type Endian int8

const (
	BigEndian    = Endian(iota) // 0
	LittleEndian                // 1
)

func (e Endian) String() string {
	switch e {
	case BigEndian:
		return "BigEndian"
	case LittleEndian:
		return "LittleEndian"
	}
	return "*ERROR-INVALID-ENDIAN-VALUE*"
}

// 4 bytes
type Header struct {
	Encoding  int8
	Endian    int8
	Reserved1 int8
	Reserved2 int8
}

//
// 20 bytes
type List struct {
	Version int32
	Flags   uint32
	Pad     uint32
}

const (
	HeaderMagic = 0x6c
)

// ensure the header is valid
func (h Header) Verify() error {
	return nil
}

// typedef struct nvpair {
// 	int32_t nvp_size;	/* size of this nvpair */
// 	int16_t	nvp_name_sz;	/* length of name string */
// 	int16_t	nvp_reserve;	/* not used */
// 	int32_t	nvp_value_elem;	/* number of elements for array types */
// 	data_type_t nvp_type;	/* type of value */
// 	/* name string */
// 	/* aligned ptr array for string arrays */
// 	/* aligned array of data for value */
// } nvpair_t;

//
// 16 bytes
type Pair struct {
	Size     int32
	NameSize int32
}

type Type uint32

// nvlist data types
const (
	Unknown = Type(iota) // 0
	Boolean              // 1
	Byte                 // 2
	Int16                // 3
	Uint16               // 4
	Int32                // 5
	Uint32               // 6
	Int64                // 7
	Uint64               // 8
	String               // 9
	ByteArray
	Int16Array
	Uint16Array
	Int32Array
	Uint32Array
	Int64Array
	Uint64Array
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
)

func (t Type) String() string {
	switch t {
	case Unknown:
		return "*UNKKNOWN-TYPE0*"
	case Boolean:
		return "Boolean"
	case Byte:
		return "Byte"
	case Int16:
		return "Int16"
	case Uint16:
		return "Uint16"
	case Int32:
		return "Int32"
	case Uint32:
		return "Uint32"
	case Int64:
		return "Int64"
	case Uint64:
		return "Uint64"
	case String:
		return "String"
	case ByteArray:
		return "ByteArray"
	case Int16Array:
		return "Int16Array"
	case Uint16Array:
		return "Uint16Array"
	case Int32Array:
		return "Int32Array"
	case Uint32Array:
		return "Uint32Array"
	case Int64Array:
		return "Int64Array"
	case Uint64Array:
		return "Uint64Array"
	case StringArray:
		return "StringArray"
	case HRTime:
		return "HRTime"
	case NVList:
		return "NVList"
	case NVListArray:
		return "NVListArray"
	case BooleanValue:
		return "BooleanValue"
	case Int8:
		return "Int8"
	case Uint8:
		return "Uint8"
	case BooleanArray:
		return "BooleanArray"
	case Int8Array:
		return "Int8Array"
	case Uint8Array:
		return "Uint8Array"
	}

	return "*UNKNOWN-TYPE*"
}

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
