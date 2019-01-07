package nvlist

// 4 bytes
type Header struct {
	Encoding  Encoding
	Endian    Endian
	Reserved1 int8
	Reserved2 int8
}
