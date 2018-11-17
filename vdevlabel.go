package zfs

import (
	"fmt"
	"unsafe"
)

type VdevLabel struct {
	BlankSpace      [8 << 10]byte   // 8k blank to accommodate os data
	BootBlockHeader [8 << 10]byte   // 8k reserved blank space
	NVP             [112 << 10]byte // XDR encoded  name value pairs
	UberBlockBuf    [128 << 10]byte // uber block array
}

func (vdl *VdevLabel) ActiveUberBlock() (*UberBlock, error) {
	ubs := vdl.UberBlocks()

	u := UberBlock{Timestamp: 0}
	// find uber block with latest timestamp
	for i := range ubs {
		if ubs[i].Timestamp > u.Timestamp {
			u = ubs[i]
		}
	}

	if u.Timestamp == 0 {
		return nil, fmt.Errorf("could not find a valid uberblock")
	}

	return &u, nil
}

func (vdl *VdevLabel) UberBlocks() []UberBlock {
	p := uintptr(unsafe.Pointer(&vdl.UberBlockBuf))
	// FIXME: this is a magic number.  hard coded 4k uber block size.  this
	// matches what i've observed but conflicts with the documentation.
	ubs := uintptr(4096)
	nrecords := uintptr((128 << 10) / ubs)

	rc := []UberBlock{}
	for i := uintptr(0); i < nrecords; i++ {
		ub := (*UberBlock)(unsafe.Pointer(p + (i * ubs)))
		if ub.Magic != 0xbab10c {
			// invalid uber block
			continue
		}
		rc = append(rc, *ub)
	}
	return rc
}
