package zfs

import (
	"fmt"
	"io"
	"log"
)

// SPA data represented as a adata virtual addresses (DVA) - 128bytes
type BlockPointer struct {
	Vdevs                 [3]VdevOffset     //  48 bytes
	Props                 BlockPointerProps //   8 bytes Props
	Padding               [2]uint64         //	16 bytes padding
	BirthTransactionGroup uint64            //   8 bytes transaction group for which this block pointer was allocated.
	Birth                 uint64            //   8 bytes transaction group for which this block pointer was allocated.
	FillCount             uint64            //   8 bytes number of non-zero block pointers under this block pointer
	ChecksumList          [4]uint64         //  32 bytes
}

func (bp BlockPointer) String() string {
	return fmt.Sprintf("BirthTransactionGroup %d", bp.BirthTransactionGroup) +
		fmt.Sprintf("Birth: %d", bp.Birth) +
		fmt.Sprintf("Fill Count: %d", bp.FillCount) +
		fmt.Sprintf("Checksum: %d\n", bp.ChecksumList) +
		fmt.Sprintf("  Props = %d (%b)\n", bp.Props, bp.Props) +
		fmt.Sprintf("  Props.Endian() = %s\n", bp.Props.Endian()) +
		fmt.Sprintf("  Props.Type() = %d\n", bp.Props.Type()) +
		fmt.Sprintf("  Props.Checksum() = %d (%s)\n", bp.Props.Checksum(), bp.Props.ChecksumString()) +
		fmt.Sprintf("  Props.Lsize() = %d\n", bp.Props.Lsize()) +
		fmt.Sprintf("  Props.Psize() = %d\n", bp.Props.Psize()) +
		fmt.Sprintf("  Props.Embedded() = %v\n", bp.Props.Embedded()) +
		fmt.Sprintf("  Props.Compression() = %d (%s)\n", bp.Props.Compression(), bp.Props.CompressionString())
}

func (bp BlockPointer) GetRootBlock(r io.ReadSeeker) (*DnodePhys, error) {
	log.Printf("offset = %d", bp.Vdevs[1].Block())
	if _, err := r.Seek(int64(bp.Vdevs[1].Offset), io.SeekStart); err != nil {
		log.Printf("err: %v", err)
		return nil, err
	}
	cmp := bp.Props.Compression()
	newr := cmp.NewReader(r)
	log.Printf("%v %T", newr, newr)
	return bp.Vdevs[0].ReadDnode(newr)
}
