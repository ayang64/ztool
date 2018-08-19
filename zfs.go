/*

	map to beginning of directory hierarchy:

	VdevLabel
		-> RootBlockPointer
			-> PhysMetaNote


*/

package zfs

import (
	"fmt"
	"io"
	"log"
	"unsafe"
)

type VdevOffset struct {
	VDEV   uint32 // id of vdev
	Size   uint32 // first byte is GRID (whatever that means) and remaning 3 bytes are ASIZE (allocated size)
	Offset uint64 // first bit is G (whatever that is) and the remainder is the offset into the vdev
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
	// Gang blocks are identified by the “G” bit.

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
		return "*ERROR-BELOW_RANGE*"
	}
	if int(zct) > len(vals)-1 {
		return "*ERROR-ABOVE-RANGE*"
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

func (bpp BlockPointerProps) Compression() uint8 {
	return uint8(bpp>>32) & 0x7f
}

func (bpp BlockPointerProps) Type() uint8 {
	return uint8(bpp>>48) & 0xff
}

func (bpp BlockPointerProps) Checksum() uint8 {
	return uint8(bpp>>40) & 0xff
}

func (bpp BlockPointerProps) Endian() string {
	return []string{"BigEndian", "LittleEndian"}[(bpp >> 63)]
}

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

// struct uberblock {
// 	/*   8 */	uint64_t	ub_magic;			/* UBERBLOCK_MAGIC		*/
// 	/*   8 */ uint64_t	ub_version;		/* SPA_VERSION			*/
// 	/*   8 */ uint64_t	ub_txg;				/* txg of last sync		*/
// 	/*   8 */ uint64_t	ub_guid_sum;	/* sum of all vdev guids	*/
// 	/*   8 */ uint64_t	ub_timestamp;	/* UTC time of last sync	*/
// 	/* 128 */ blkptr_t	ub_rootbp;		/* MOS objset_phys_t		*/
// };
//
// 168 bytes

// UberBlock -- 1024bits - 128bytes
// comments cribbed from /usr/src/sys/cddl/boot/zfs/zfssimpl.h
type UberBlock struct {
	Magic            uint64       // magic 0x00babl0c (oo-ba-bloc!)
	Version          uint64       // SPA Version
	TransactionGroup uint64       // transaction group of last sync
	GuidSum          uint64       // sum of all vdev guids
	Timestamp        uint64       // time of last sync
	RootBP           BlockPointer // mos objset_phys_t
	SoftwareVersion  uint64
	PAD              [3]uint64
	CheckpointTx     uint64
}

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

	for idx := range rc {
		log.Printf("rc[%d].Magic = %x", idx, rc[idx].Magic)
	}
	return rc
}

type ZfsVdev struct {
	Back io.ReadSeeker
	VDL  VdevLabel
}

func (z *ZfsVdev) Read() {
}
