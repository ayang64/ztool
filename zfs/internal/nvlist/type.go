package nvlist

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

type Type int32

// nvlist data types
const (
	DontCare = Type(iota - 1)
	Unknown
	Boolean
	Byte
	Int16
	Uint16
	Int32
	Uint32
	Int64
	Uint64
	String
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
