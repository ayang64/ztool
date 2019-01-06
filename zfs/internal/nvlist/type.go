package nvlist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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
	return "*ERROR-INVALID-ENDIAN-VALUE*"
}

type Encoding uint8

func (e Encoding) String() string {
	if e == EncodingNative {
		return "EncodingNative"
	}

	if e == EncodingXDR {

		return "EncodingXDR"
	}

	return fmt.Sprintf("*UNKNOWN-ENCODING-%02x*", uint8(e))
}

const (
	EncodingNative = Encoding(iota)
	EncodingXDR
)

// 4 bytes
type Header struct {
	Encoding  Encoding
	Endian    Endian
	Reserved1 int8
	Reserved2 int8
}

//
// 20 bytes
type List struct {
	Version int32
	Flags   uint32
}

type Scanner struct {
	byteOrder        binary.ByteOrder
	r                io.Reader
	header           Header
	list             List
	pair             Pair
	fieldName        string
	fieldType        Type
	fieldNumElements int
	value            interface{}
	err              error
	bytes            []byte
}

func (s *Scanner) ReadNumElements(r io.Reader) (int32, error) {
	var rc int32
	if err := binary.Read(r, s.byteOrder, &rc); err != nil {
		return -1, err
	}
	return rc, nil
}

func (s *Scanner) ReadType(r io.Reader) (Type, error) {
	var rc Type
	if err := binary.Read(r, s.byteOrder, &rc); err != nil {
		return -1, err
	}
	return rc, nil
}

func (s *Scanner) Bytes() []byte {
	return s.bytes
}

func (s *Scanner) ValueString() string {
	switch v := s.Value().(type) {
	case uint64:
		log.Printf("v = %d", v)
		return fmt.Sprintf("%d", v)
	case string:
		return v
	}
	return "*unrepresentable*"
}

func (s *Scanner) ReadString(r io.Reader) (string, error) {
	// read 4 byte length
	var length int32

	if err := binary.Read(r, s.byteOrder, &length); err != nil {
		return "", err
	}

	align4 := func(i int32) int32 {
		return (i + 3) & ^3
	}

	str := make([]byte, align4(length))

	if err := binary.Read(r, s.byteOrder, str); err != nil {
		return "", err
	}

	return string(str[:length]), nil
}

func (s *Scanner) Name() string {
	return s.fieldName
}

func (s *Scanner) Value() interface{} {
	return s.value
}

func (s *Scanner) Type() Type {
	return s.fieldType
}

func (s *Scanner) NumElements() int {
	return s.fieldNumElements
}

func (s *Scanner) Error() error {
	if s.err != nil {
		return s.err
	}
	return nil
}

func (s *Scanner) FieldSize() int {
	return int(s.pair.Size)
}

func (s *Scanner) Next() bool {
	// Do not continue if the scanner is in an errored state.
	if s.err != nil {
		return false
	}

	if s.err = binary.Read(s.r, s.byteOrder, &s.pair); s.err != nil {
		return false
	}

	if s.pair.Size == 0 && s.pair.DecodedSize == 0 {
		return false
	}

	// read entire record into a byte slice
	record := make([]byte, s.pair.Size-8)
	if s.err = binary.Read(s.r, s.byteOrder, record); s.err != nil {
		return false
	}

	// lets read from the remainding bytes
	br := bytes.NewReader(record)

	// read the name of the field
	name, err := s.ReadString(br)

	if err != nil {
		s.err = err
		return false
	}

	s.fieldName = name

	typ, err := s.ReadType(br)

	if err != nil {
		s.err = err
		return false
	}

	s.fieldType = typ

	nelements, err := s.ReadNumElements(br)

	if err != nil {
		s.err = err
		return false
	}

	s.fieldNumElements = int(nelements)

	log.Printf("about to read from %#v", record)
	value, err := s.ReadValue(br, s.fieldType)

	if err != nil {
		s.err = err
		return false
	}

	s.value = value

	return true
}

func (s *Scanner) NewSubScanner(r io.Reader) (rc *Scanner) {
	rc = &Scanner{
		r:         r,
		byteOrder: s.byteOrder,
	}

	// if err := binary.Read(r, rc.byteOrder, &rc.list); err != nil {
	if err := binary.Read(r, s.byteOrder, &rc.list); err != nil {
		rc.err = err
		return
	}

	return

}

func NewScanner(r io.Reader) (rc *Scanner) {
	rc = &Scanner{
		r: r,
	}

	// at the moment, byte order doesn't matter.
	if err := binary.Read(r, binary.BigEndian, &rc.header); err != nil {
		rc.err = err
		return
	}

	rc.byteOrder = func() binary.ByteOrder {
		if rc.header.Endian == BigEndian {
			return binary.BigEndian
		}
		return binary.LittleEndian
	}()

	log.Printf("HEADER: %#v", rc.header)
	log.Printf("ENDIAN: %s", rc.header.Endian)
	log.Printf("ENCODING: %s", rc.header.Encoding)

	// if err := binary.Read(r, rc.byteOrder, &rc.list); err != nil {
	if err := binary.Read(r, rc.byteOrder, &rc.list); err != nil {
		rc.err = err
		return
	}

	log.Printf("LIST: %#v", rc.list)

	return
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
	Size        int32
	DecodedSize int32
}

type Type int32

// nvlist data types
const (
	DontCare = Type(iota - 1) // -1
	Unknown                   // 0
	Boolean                   // 1
	Byte                      // 2
	Int16                     // 3
	Uint16                    // 4
	Int32                     // 5
	Uint32                    // 6
	Int64                     // 7
	Uint64                    // 8
	String                    // 9
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

func (s *Scanner) ReadValue(r io.Reader, t Type) (interface{}, error) {
	switch t {
	case DontCare:
	case Unknown:
	case Boolean:
		return true, nil

	case Byte:
	case Int16:
	case Uint16:
	case Int32:
	case Uint32:
	case Int64:
	case Uint64:
		var rc uint64
		if err := binary.Read(r, s.byteOrder, &rc); err != nil {
			return nil, err
		}
		return rc, nil

	case String:
		var length int32
		if err := binary.Read(r, s.byteOrder, &length); err != nil {
			return "", err
		}

		align4 := func(i int32) int32 {
			return (i + 3) & ^3
		}

		str := make([]byte, align4(length))

		if err := binary.Read(r, s.byteOrder, str); err != nil {
			return "", err
		}

		return string(str[:length]), nil

	case ByteArray:
	case Int16Array:
	case Uint16Array:
	case Int32Array:
	case Uint32Array:
	case Int64Array:
	case Uint64Array:
	case StringArray:
	case HRTime:
	case NVList:
		return s.ReadSub(r)

	case NVListArray:
		rc := make([]map[string]interface{}, s.NumElements())
		for i := 0; i < s.NumElements(); i++ {
			v, err := s.ReadSub(r)

			if err != nil {
				log.Fatal(err)
			}
			rc = append(rc, v)
		}
		return rc, nil

	case BooleanValue:
	case Int8:
	case Uint8:
	case BooleanArray:
	case Int8Array:
	case Uint8Array:
	}

	return nil, fmt.Errorf("unknown type %q", t)
}

func (s *Scanner) ReadSub(r io.Reader) (map[string]interface{}, error) {
	rc := make(map[string]interface{})
	scn := s.NewSubScanner(r)
	if err := scn.Error(); err != nil {
		return nil, err
	}
	for scn.Next() {
		log.Printf("name: %s, value type: %s", scn.Name(), scn.Type())
		rc[scn.Name()] = scn.Value()
		log.Printf("  NAME: %s", scn.Name())
		log.Printf("  TYPE: %s", scn.Type())
	}

	if err := scn.Error(); err != nil {
		log.Printf("errored field: %s (%d)", scn.Name(), scn.Type())
		log.Printf("errored type: %s", scn.Type())
		log.Printf("errored num elements: %d", scn.NumElements())
		return nil, err
	}

	return rc, nil
}

func ReadFull(r io.Reader) (map[string]interface{}, error) {
	rc := make(map[string]interface{})
	scn := NewScanner(r)
	if err := scn.Error(); err != nil {
		return nil, err
	}

	log.Printf("scn: %#v", scn)

	for scn.Next() {
		log.Printf("name: %s, value type: %s", scn.Name(), scn.Type())
		rc[scn.Name()] = scn.Value()
		log.Printf("  NAME: %s", scn.Name())
	}

	if err := scn.Error(); err != nil {
		log.Printf("errored field: %s", scn.Name())
		log.Printf("errored type: %s", scn.Type())
		log.Printf("errored num elements: %d", scn.NumElements())
		return nil, err
	}

	return rc, nil
}

func (t Type) Size() int32 {
	sizes := map[Type]int32{
		DontCare:     -1,
		Unknown:      -1,
		Boolean:      0, // unknown size
		Byte:         1,
		Int16:        2,
		Uint16:       2,
		Int32:        4,
		Uint32:       4,
		Int64:        8,
		Uint64:       8,
		String:       1,
		ByteArray:    -1,
		Int16Array:   -1,
		Uint16Array:  -1,
		Int32Array:   -1,
		Uint32Array:  -1,
		Int64Array:   -1,
		Uint64Array:  -1,
		StringArray:  -1,
		HRTime:       -1,
		NVList:       -1,
		NVListArray:  -1,
		BooleanValue: -1,
		Int8:         1,
		Uint8:        1,
		BooleanArray: -1,
		Int8Array:    -1,
		Uint8Array:   -1,
	}

	size, exists := sizes[t]

	if exists == false {
		return -1
	}

	return size
}

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
