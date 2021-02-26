package nvlist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// Scanner provides a convenient way to read an XDR encode nvlist from a ZFS
// volume.  The Scanner type encapsulates all of the context related to
// iterating over nvlist entries.
type Scanner struct {
	r                io.Reader // io.Reader for reading scanned data.
	withheader       bool
	byteOrder        binary.ByteOrder // Byte order of the values being read.
	header           Header           // Non-repeating portion of nvlist that contains the encoding type and byte order of the data in the list being scanned.
	list             ListMeta         // Non-repeating data encoding the nvlist version and any flags.
	pair             Pair
	fieldName        string
	fieldType        Type
	fieldNumElements int
	value            interface{}
	err              error
	bytes            []byte
}

func WithByteOrder(o binary.ByteOrder) func(*Scanner) error {
	return func(s *Scanner) error {
		s.byteOrder = o
		return nil
	}
}

func WithoutHeader() func(*Scanner) error {
	return func(s *Scanner) error {
		log.Printf("setting withheader to false.")
		s.withheader = false
		return nil
	}
}

func NewScanner(r io.Reader, opts ...func(*Scanner) error) (rc *Scanner) {
	rc = &Scanner{
		r:          r,
		withheader: true,
	}

	for _, opt := range opts {
		opt(rc)
	}

	// at the moment, byte order doesn't matter.
	if rc.withheader {
		log.Printf("scanning header.")
		if err := binary.Read(r, binary.BigEndian, &rc.header); err != nil {
			rc.err = err
			return
		}
	} else {
		log.Printf("skipping header.")
	}

	rc.byteOrder = func() binary.ByteOrder {
		if rc.header.Endian == BigEndian {
			return binary.BigEndian
		}
		return binary.LittleEndian
	}()

	if err := binary.Read(r, rc.byteOrder, &rc.list); err != nil {
		rc.err = err
		return
	}
	return
}

func (s *Scanner) ReadValue(r io.Reader, t Type) (interface{}, error) {
	f := s.readValueFunc(r, t)

	if f == nil {
		return nil, fmt.Errorf("no conversion function for type %q", t)
	}

	return f()
}

func (s *Scanner) readValueFunc(r io.Reader, t Type) func() (interface{}, error) {
	m := map[Type]func() (interface{}, error){
		DontCare: nil,
		Unknown:  nil,
		Boolean:  func() (interface{}, error) { return true, nil },
		Byte:     nil,
		Int16:    nil,
		Uint16:   nil,
		Int32:    nil,
		Uint32:   nil,
		Int64:    nil,
		Uint64: func() (interface{}, error) {
			var rc uint64
			if err := binary.Read(r, s.byteOrder, &rc); err != nil {
				return nil, err
			}
			return rc, nil
		},
		String: func() (interface{}, error) {
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
		},
		ByteArray:   nil,
		Int16Array:  nil,
		Uint16Array: nil,
		Int32Array:  nil,
		Uint32Array: nil,
		Int64Array:  nil,
		Uint64Array: nil,
		StringArray: nil,
		HRTime:      nil,
		NVList:      func() (interface{}, error) { return s.ReadSub(r) },
		NVListArray: func() (interface{}, error) {
			rc := make([]List, 0, s.NumElements())
			for i := 0; i < s.NumElements(); i++ {
				v, err := s.ReadSub(r)

				if err != nil {
					log.Fatal(err)
				}
				rc = append(rc, v)
			}
			return rc, nil
		},
		BooleanValue: nil,
		Int8:         nil,
		Uint8:        nil,
		BooleanArray: nil,
		Int8Array:    nil,
		Uint8Array:   nil,
	}

	if f, found := m[t]; found {
		return f
	}

	return nil
}

func (s *Scanner) ReadSub(r io.Reader) (List, error) {
	rc := make(List)
	scn := s.NewSubScanner(r)
	if err := scn.Error(); err != nil {
		return nil, err
	}
	for scn.Next() {
		rc[scn.Name()] = scn.Value()
	}

	if err := scn.Error(); err != nil {
		return nil, err
	}

	return rc, nil
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
