package zfs

import (
	"io"
	"unsafe"
)

type VdevOffset struct {
	VDEV   uint32 // id of vdev
	Size   uint32 // first byte is GRID (whatever that means) and remaning 3 bytes are ASIZE (allocated size)
	Offset uint64 // first bit is G (whatever that is) and the remainder is the offset into the vdev
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

// SPA data represented as a adata virtual addresses (DVA) - 128bytes
type BlockPointer struct {
	Vdevs [3]VdevOffset // 48 bytes
	Props struct {
		Level       uint8              // 1
		Type        uint8              // 1
		Checksum    uint8              // 1
		Compression ZfsCompressionType // 1
		PSize       uint16             // 2
		LSize       uint16             // 2
	}
	Padding               [2]uint64 // 	16 bytes padding
	BirthTransactionGroup uint64    //  8 bytes transaction group for which this block pointer was allocated.
	Birth                 int64     //  8 bytes transaction group for which this block pointer was allocated.
	FillCount             uint64    //  8 bytes number of non-zero block pointers under this block pointer
	ChecksumList          [4]uint64 // 32 bytes
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
}

type VdevLabel struct {
	BlankSpace      [8 << 10]byte   // 8k blank to accommodate os data
	BootBlockHeader [8 << 10]byte   // 8k reserved blank space
	NVP             [112 << 10]byte // XDR encoded  name value pairs
	UberBlockBuf    [128 << 10]byte // uber block array
}

func (vdl *VdevLabel) UberBlock() []UberBlock {
	ub := *(*[(128 << 10) / (168 * 8)]UberBlock)(unsafe.Pointer(&vdl.UberBlockBuf))
	return ub[:]
}

type ZfsVdev struct {
	Back io.ReadSeeker
	VDL  VdevLabel
}

func (z *ZfsVdev) Read() {
}
