package zle

import (
	"testing"
)

func TestZLE(t *testing.T) {
	// srclen = 25, dcmrem = 0
	src := []byte{0, 0, 0, 0, 12, 0, 0, 0, 0, 0, -1, 0, 0, 2, 8, 10, 0, 0, 0, -128, 0, 0, 0, 0, 8}
	cmp := []byte{11, 0, 12, 12, 0, -1, 9, 2, 2, 8, 10, 10, 0, -128, 11, 0, 8}

}
