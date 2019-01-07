package nvlist

import "fmt"

type Encoding uint8

const (
	EncodingNative = Encoding(iota)
	EncodingXDR
)

func (e Encoding) String() string {
	if e == EncodingNative {
		return "EncodingNative"
	}

	if e == EncodingXDR {

		return "EncodingXDR"
	}

	return fmt.Sprintf("*UNKNOWN-ENCODING-%02x*", uint8(e))
}
