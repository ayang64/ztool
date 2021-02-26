package zfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// typedef struct blkptr {
// 	dva_t		blk_dva[SPA_DVAS_PER_BP]; /* Data Virtual Addresses */
// 	uint64_t	blk_prop;	/* size, compression, type, etc	    */
// 	uint64_t	blk_pad[2];	/* Extra space for the future	    */
// 	uint64_t	blk_phys_birth;	/* txg when block was allocated	    */
// 	uint64_t	blk_birth;	/* transaction group at birth	    */
// 	uint64_t	blk_fill;	/* fill count			    */
// 	zio_cksum_t	blk_cksum;	/* 256-bit checksum		    */
// } blkptr_t;

// SPA data represented as a adata virtual addresses (DVA) - 128bytes
type BlockPointer struct {
	Vdevs                 [3]DVA            //  48 bytes
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
		fmt.Sprintf("  Props.Level() = %d\n", bp.Props.Level()) +
		fmt.Sprintf("  Props.Type() = %d\n", bp.Props.Type()) +
		fmt.Sprintf("  Props.Checksum() = %d (%s)\n", bp.Props.Checksum(), bp.Props.ChecksumString()) +
		fmt.Sprintf("  Props.Lsize() = %d\n", bp.Props.Lsize()) +
		fmt.Sprintf("  Props.Psize() = %d\n", bp.Props.Psize()) +
		fmt.Sprintf("  Props.Embedded() = %v\n", bp.Props.Embedded()) +
		fmt.Sprintf("  Props.Compression() = %d (%s)\n", bp.Props.Compression(), bp.Props.CompressionString())
}

func (bp *BlockPointer) GetDnode(r io.ReadSeeker) (*DnodePhys, error) {
	vdev := 0
	log.Printf("offset = %d", bp.Vdevs[vdev].Block())

	if _, err := r.Seek(int64(bp.Vdevs[vdev].Block()), io.SeekStart); err != nil {
		log.Printf("err: %v", err)
		return nil, err
	}

	cmp := bp.Props.Compression()

	pbuf := make([]byte, bp.Props.Psize())
	lbuf := make([]byte, bp.Props.Lsize())

	binary.Read(r, binary.LittleEndian, pbuf)
	cmp.Decompress(lbuf, pbuf)
	log.Printf("pbuf: %v", pbuf)
	log.Printf("lbuf: %v", lbuf)

	return bp.Vdevs[vdev].ReadDnode(bytes.NewReader(lbuf))
}
