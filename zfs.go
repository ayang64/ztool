/*
	map to beginning of directory hierarchy:

	VdevLabel
		-> RootBlockPointer
			-> PhysMetaNote
*/

package zfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
	"unsafe"

	"github.com/pierrec/lz4"
)

type VdevOffset struct {
	VDEV   uint32 // id of vdev
	Size   uint32 // first byte is GRID (whatever that means) and remaning 3 bytes are ASIZE (allocated size)
	Offset uint64 // first bit is G (whatever that is) and the remainder is the offset into the vdev
}

func (vdo *VdevOffset) ReadDnode(r io.Reader) (*DnodePhys, error) {
	dn := DnodePhys{}

	if err := binary.Read(r, binary.LittleEndian, &dn); err != nil {
		log.Fatal(err)
	}

	log.Printf("dn = %#v", dn)
	return &dn, nil
}

func (vdo *VdevOffset) Asize() int {
	return int(vdo.Size) & 0x0ffffff
}

func (vdo *VdevOffset) Block() uint64 {
	// ZFS talks about data in terms of 512byte blocks. the actual location is
	// 4mb + (512 * offset) the shift gets rid of the G bit which is stored in
	// the high order bit of vdo.Offset.
	return (vdo.Offset << 9) + 0x400000
}

func (vdo *VdevOffset) Gang() bool {
	// From the "ZFS On Disk Format" document:
	//
	// A gang block is a block whose contents contain block pointers. Gang blocks
	// are used when the amount of space requested is not available in a
	// contiguous block. In a situation of this kind, several smaller blocks will
	// be allocated (totaling up to the size requested) and a gang block will be
	// created to contain the block pointers for the allocated blocks. A pointer
	// to this gang block is returned to the requester, giving the requester the
	// perception of a single block.
	//
	// Gang blocks are identified by the G bit

	// we do some simple bit shifting to return a bool representing
	// the most significant bit in our offset.
	return vdo.Offset&(1<<63) != 0
}

// Cribbed from zfsimpl.h
//
// 	uint8_t dn_type;		/* dmu_object_type_t */
// 	uint8_t dn_indblkshift;		/* ln2(indirect block size) */
// 	uint8_t dn_nlevels;		/* 1=dn_blkptr->data blocks */
// 	uint8_t dn_nblkptr;		/* length of dn_blkptr */
// 	uint8_t dn_bonustype;		/* type of data in bonus buffer */
// 	uint8_t	dn_checksum;		/* ZIO_CHECKSUM type */
// 	uint8_t	dn_compress;		/* ZIO_COMPRESS type */
// 	uint8_t dn_flags;		/* DNODE_FLAG_* */
// 	uint16_t dn_datablkszsec;	/* data block size in 512b sectors */
// 	uint16_t dn_bonuslen;		/* length of dn_bonus */
// 	uint8_t dn_extra_slots;		/* # of subsequent slots consumed */
// 	uint8_t dn_pad2[3];
// 	/* accounting is protected by dn_dirty_mtx */
// 	uint64_t dn_maxblkid;		/* largest allocated block ID */
// 	uint64_t dn_used;		/* bytes (or sectors) of disk space */
// 	uint64_t dn_pad3[4];
// 	blkptr_t dn_blkptr[1+DN_OLD_MAX_BONUSLEN/sizeof (blkptr_t)]; // 3 in the docs
// 	union {
// 		blkptr_t dn_blkptr[1+DN_OLD_MAX_BONUSLEN/sizeof (blkptr_t)];
// 		struct {
// 			blkptr_t __dn_ignore1;
// 			uint8_t dn_bonus[DN_OLD_MAX_BONUSLEN];
// 		};
// 		struct {
// 			blkptr_t __dn_ignore2;
// 			uint8_t __dn_ignore3[DN_OLD_MAX_BONUSLEN -
// 			    sizeof (blkptr_t)];
// 			blkptr_t dn_spill;
// 		};
// 	};

type DnodePhys struct {
	Type               uint8              // dn type
	IndirectBlockShift uint8              // ln2(indirect block size) -- indirect_block_size^2 = size of block?
	IndirectionLevels  uint8              // 1=dn_blkptr->data blocks
	BlockPointerLength uint8              // length of dn_blkptr
	BonusType          uint8              // type of data in bonus buffer
	Checksum           uint8              // ZIO_CHECKSUM type
	Compress           ZfsCompressionType // ZIO_COMPRESS type
	Flags              uint8              // DNODE_FLAG_*
	DataBlockSize      uint16             // data block size in 512b sectors
	BonusLength        uint16             // length of dn_bonus
	ExtraSlots         uint8              // # of subsequent slots consumed
	Pad2               [3]uint8           // 3 bytes padding
	MaxBlockID         uint64             // largest allocated block ID
	Used               uint64             // bytes (or sectors) of disk space
	Pad3               [4]uint64          // 24 bytes padding
	BlockPointer       [3]BlockPointer    //
	BONUS              [8]uint64          // not sure what to do with this yet.
}

type ZfsCompressionType uint8

func (zct ZfsCompressionType) NewReader(r io.ReadSeeker) io.Reader {
	switch zct.String() {
	case "ZIO_COMPRESS_INHERIT":
		return r
	case "ZIO_COMPRESS_ON":
		return r
	case "ZIO_COMPRESS_LZ4":
		return lz4.NewReader(r)
	case "ZIO_COMPRESS_ZLE":
		return r
	default:
		return r
	}
}

func (zct ZfsCompressionType) String() string {
	vals := []string{
		"ZIO_COMPRESS_INHERIT",
		"ZIO_COMPRESS_ON",
		"ZIO_COMPRESS_OFF",
		"ZIO_COMPRESS_LZJB",
		"ZIO_COMPRESS_EMPTY",
		"ZIO_COMPRESS_GZIP_1",
		"ZIO_COMPRESS_GZIP_2",
		"ZIO_COMPRESS_GZIP_3",
		"ZIO_COMPRESS_GZIP_4",
		"ZIO_COMPRESS_GZIP_5",
		"ZIO_COMPRESS_GZIP_6",
		"ZIO_COMPRESS_GZIP_7",
		"ZIO_COMPRESS_GZIP_8",
		"ZIO_COMPRESS_GZIP_9",
		"ZIO_COMPRESS_ZLE",
		"ZIO_COMPRESS_LZ4",
		"ZIO_COMPRESS_FUNCTIONS",
	}
	if zct < 0 {
		return fmt.Sprintf("*ERROR-%03d-BELOW-RANGE*", zct)
	}
	if int(zct) > len(vals)-1 {
		return fmt.Sprintf("*ERROR-%03d-ABOVE-RANGE*", zct)
	}
	return vals[zct]
}

// typedef struct blkptr {
// 	/* 48 */	dva_t		blk_dva[SPA_DVAS_PER_BP]; /* Data Virtual Addresses */
// 	/*  8 */	uint64_t	blk_prop;				/* size, compression, type, etc	    */
// 	/* 16 */	uint64_t	blk_pad[2];			/* Extra space for the future	    */
// 	/*  8 */	uint64_t	blk_phys_birth;	/* txg when block was allocated	    */
// 	/*  8 */	uint64_t	blk_birth;			/* transaction group at birth	    */
// 	/*  8 */	uint64_t	blk_fill;				/* fill count			    */
// 	/* 32 */	zio_cksum_t	blk_cksum;		/* 256-bit checksum		    */
// } blkptr_t;
//
// 128 bytes

type BlockPointerProps uint64

/*
{
	Level       uint8              // 1
	Type        uint8              // 1
	Checksum    uint8              // 1
	Compression ZfsCompressionType // 1
	PSize       uint16             // 2
	LSize       uint16             // 2
}
*/

func (bpp BlockPointerProps) Embedded() bool {
	return (uint64(bpp>>39) & 0x01) == 1
}

func (bpp BlockPointerProps) Lsize() uint8 {
	return uint8(bpp & 0xff)
}

func (bpp BlockPointerProps) Psize() uint8 {
	return uint8((bpp >> 8) & 0xff)
}

func (bpp BlockPointerProps) Compression() ZfsCompressionType {
	return ZfsCompressionType(uint8(bpp>>32) & 0x7f)
}
func (bpp BlockPointerProps) CompressionString() string {
	return bpp.Compression().String()
}

func (bpp BlockPointerProps) Type() uint8 {
	return uint8(bpp>>48) & 0xff
}

func (bpp BlockPointerProps) Checksum() uint8 {
	return uint8(bpp>>40) & 0xff
}

func (bpp BlockPointerProps) ChecksumString() string {
	chk := []string{
		"ZIO_CHECKSUM_INHERIT",
		"ZIO_CHECKSUM_ON",
		"ZIO_CHECKSUM_OFF",
		"ZIO_CHECKSUM_LABEL",
		"ZIO_CHECKSUM_GANG_HEADER",
		"ZIO_CHECKSUM_ZILOG",
		"ZIO_CHECKSUM_FLETCHER_2",
		"ZIO_CHECKSUM_FLETCHER_4",
		"ZIO_CHECKSUM_SHA256",
		"ZIO_CHECKSUM_ZILOG2",
		"ZIO_CHECKSUM_NOPARITY",
		"ZIO_CHECKSUM_SHA512",
		"ZIO_CHECKSUM_SKEIN",
		"ZIO_CHECKSUM_EDONR",
		"ZIO_CHECKSUM_FUNCTIONS",
	}
	c := int(bpp.Checksum())
	if c < 0 {
		return "*ERROR-CHECKSUM-UNDER-BOUNDS*"
	}
	if c > len(chk)-1 {
		return "*ERROR-CHECKSUM-OVER-BOUNDS*"
	}
	return chk[c]
}

func (bpp BlockPointerProps) Endian() string {
	return []string{"BigEndian", "LittleEndian"}[(bpp >> 63)]
}

type ZfsVdev struct {
	Back io.ReadSeeker
	VDL  VdevLabel
}

func (z *ZfsVdev) Read() {
}
