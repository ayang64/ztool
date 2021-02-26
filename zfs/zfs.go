// Copyright 2018 Ayan George.
// All rights reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	map to beginning of directory hierarchy:

	VdevLabel
	  -> UberBlock
		 -> RootBlockPointer
			-> PhysMetaNote
*/

package zfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	// "github.com/pierrec/lz4"
	lz4 "github.com/bkaradzic/go-lz4"
)

type DVA struct {
	VDEV   uint32 // id of vdev
	Size   uint32 // first byte is GRID (whatever that means) and remaning 3 bytes are ASIZE (allocated size)
	Offset uint64 // first bit is G (whatever that is) and the remainder is the offset into the vdev
}

func (dva *DVA) ReadDnode(r io.Reader) (*DnodePhys, error) {
	dn := DnodePhys{}

	if err := binary.Read(r, binary.LittleEndian, &dn); err != nil {
		log.Fatal(err)
	}

	log.Printf("dn = %#v", dn)
	return &dn, nil
}

func (dva *DVA) Asize() int {
	return int(dva.Size) & 0x0ffffff
}

func (dva *DVA) Block() uint64 {
	// ZFS talks about data in terms of 512byte blocks. the actual location is
	// 4mb + (512 * offset) the shift gets rid of the G bit which is stored in
	// the high order bit of dva.Offset.

	mask := uint64(1 << 63)
	offs := dva.Offset &^ mask

	log.Printf("mask: %064b", mask)
	log.Printf("offs: %064b", dva.Offset)
	log.Printf("valu: %064b", offs)
	log.Printf("offs: %064b", offs<<9)
	return (offs << 9) + 0x400000
}

func (dva *DVA) Gang() bool {
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
	return dva.Offset&(1<<63) != 0
}

//typedef enum dmu_object_type {
//	DMU_OT_NONE,
//	/* general: */
//	DMU_OT_OBJECT_DIRECTORY,	/* ZAP */
//	DMU_OT_OBJECT_ARRAY,		/* UINT64 */
//	DMU_OT_PACKED_NVLIST,		/* UINT8 (XDR by nvlist_pack/unpack) */
//	DMU_OT_PACKED_NVLIST_SIZE,	/* UINT64 */
//	DMU_OT_BPLIST,			/* UINT64 */
//	DMU_OT_BPLIST_HDR,		/* UINT64 */
//	/* spa: */
//	DMU_OT_SPACE_MAP_HEADER,	/* UINT64 */
//	DMU_OT_SPACE_MAP,		/* UINT64 */
//	/* zil: */
//	DMU_OT_INTENT_LOG,		/* UINT64 */
//	/* dmu: */
//	DMU_OT_DNODE,			/* DNODE */
//	DMU_OT_OBJSET,			/* OBJSET */
//	/* dsl: */
//	DMU_OT_DSL_DIR,			/* UINT64 */
//	DMU_OT_DSL_DIR_CHILD_MAP,	/* ZAP */
//	DMU_OT_DSL_DS_SNAP_MAP,		/* ZAP */
//	DMU_OT_DSL_PROPS,		/* ZAP */
//	DMU_OT_DSL_DATASET,		/* UINT64 */
//	/* zpl: */
//	DMU_OT_ZNODE,			/* ZNODE */
//	DMU_OT_OLDACL,			/* Old ACL */
//	DMU_OT_PLAIN_FILE_CONTENTS,	/* UINT8 */
//	DMU_OT_DIRECTORY_CONTENTS,	/* ZAP */
//	DMU_OT_MASTER_NODE,		/* ZAP */
//	DMU_OT_UNLINKED_SET,		/* ZAP */
//	/* zvol: */
//	DMU_OT_ZVOL,			/* UINT8 */
//	DMU_OT_ZVOL_PROP,		/* ZAP */
//	/* other; for testing only! */
//	DMU_OT_PLAIN_OTHER,		/* UINT8 */
//	DMU_OT_UINT64_OTHER,		/* UINT64 */
//	DMU_OT_ZAP_OTHER,		/* ZAP */
//	/* new object types: */
//	DMU_OT_ERROR_LOG,		/* ZAP */
//	DMU_OT_SPA_HISTORY,		/* UINT8 */
//	DMU_OT_SPA_HISTORY_OFFSETS,	/* spa_his_phys_t */
//	DMU_OT_POOL_PROPS,		/* ZAP */
//	DMU_OT_DSL_PERMS,		/* ZAP */
//	DMU_OT_ACL,			/* ACL */
//	DMU_OT_SYSACL,			/* SYSACL */
//	DMU_OT_FUID,			/* FUID table (Packed NVLIST UINT8) */
//	DMU_OT_FUID_SIZE,		/* FUID table size UINT64 */
//	DMU_OT_NEXT_CLONES,		/* ZAP */
//	DMU_OT_SCAN_QUEUE,		/* ZAP */
//	DMU_OT_USERGROUP_USED,		/* ZAP */
//	DMU_OT_USERGROUP_QUOTA,		/* ZAP */
//	DMU_OT_USERREFS,		/* ZAP */
//	DMU_OT_DDT_ZAP,			/* ZAP */
//	DMU_OT_DDT_STATS,		/* ZAP */
//	DMU_OT_SA,			/* System attr */
//	DMU_OT_SA_MASTER_NODE,		/* ZAP */
//	DMU_OT_SA_ATTR_REGISTRATION,	/* ZAP */
//	DMU_OT_SA_ATTR_LAYOUTS,		/* ZAP */
//	DMU_OT_SCAN_XLATE,		/* ZAP */
//	DMU_OT_DEDUP,			/* fake dedup BP from ddt_bp_create() */
//	DMU_OT_NUMTYPES,
//	/*
//	 * Names for valid types declared with DMU_OT().
//	 */
//	DMU_OTN_UINT8_DATA = DMU_OT(DMU_BSWAP_UINT8, B_FALSE),
//	DMU_OTN_UINT8_METADATA = DMU_OT(DMU_BSWAP_UINT8, B_TRUE),
//	DMU_OTN_UINT16_DATA = DMU_OT(DMU_BSWAP_UINT16, B_FALSE),
//	DMU_OTN_UINT16_METADATA = DMU_OT(DMU_BSWAP_UINT16, B_TRUE),
//	DMU_OTN_UINT32_DATA = DMU_OT(DMU_BSWAP_UINT32, B_FALSE),
//	DMU_OTN_UINT32_METADATA = DMU_OT(DMU_BSWAP_UINT32, B_TRUE),
//	DMU_OTN_UINT64_DATA = DMU_OT(DMU_BSWAP_UINT64, B_FALSE),
//	DMU_OTN_UINT64_METADATA = DMU_OT(DMU_BSWAP_UINT64, B_TRUE),
//	DMU_OTN_ZAP_DATA = DMU_OT(DMU_BSWAP_ZAP, B_FALSE),
//	DMU_OTN_ZAP_METADATA = DMU_OT(DMU_BSWAP_ZAP, B_TRUE)
//} dmu_object_type_t;

type DmuObjectType uint8

const (
	//typedef enum dmu_object_type {
	DMU_OT_NONE = DmuObjectType(iota)
	//	/* general: */
	DMU_OT_OBJECT_DIRECTORY   /* ZAP */
	DMU_OT_OBJECT_ARRAY       /* UINT64 */
	DMU_OT_PACKED_NVLIST      /* UINT8 (XDR by nvlist_pack/unpack) */
	DMU_OT_PACKED_NVLIST_SIZE /* UINT64 */
	DMU_OT_BPLIST             /* UINT64 */
	DMU_OT_BPLIST_HDR         /* UINT64 */
	//	/* spa: */
	DMU_OT_SPACE_MAP_HEADER /* UINT64 */
	DMU_OT_SPACE_MAP        /* UINT64 */
	//	/* zil: */
	DMU_OT_INTENT_LOG /* UINT64 */
	//	/* dmu: */
	DMU_OT_DNODE  /* DNODE */
	DMU_OT_OBJSET /* OBJSET */
	//	/* dsl: */
	DMU_OT_DSL_DIR           /* UINT64 */
	DMU_OT_DSL_DIR_CHILD_MAP /* ZAP */
	DMU_OT_DSL_DS_SNAP_MAP   /* ZAP */
	DMU_OT_DSL_PROPS         /* ZAP */
	DMU_OT_DSL_DATASET       /* UINT64 */
	//	/* zpl: */
	DMU_OT_ZNODE               /* ZNODE */
	DMU_OT_OLDACL              /* Old ACL */
	DMU_OT_PLAIN_FILE_CONTENTS /* UINT8 */
	DMU_OT_DIRECTORY_CONTENTS  /* ZAP */
	DMU_OT_MASTER_NODE         /* ZAP */
	DMU_OT_UNLINKED_SET        /* ZAP */
	//	/* zvol: */
	DMU_OT_ZVOL      /* UINT8 */
	DMU_OT_ZVOL_PROP /* ZAP */
	//	/* other; for testing only! */
	DMU_OT_PLAIN_OTHER  /* UINT8 */
	DMU_OT_UINT64_OTHER /* UINT64 */
	DMU_OT_ZAP_OTHER    /* ZAP */
	//	/* new object types: */
	DMU_OT_ERROR_LOG            /* ZAP */
	DMU_OT_SPA_HISTORY          /* UINT8 */
	DMU_OT_SPA_HISTORY_OFFSETS  /* spa_his_phys_t */
	DMU_OT_POOL_PROPS           /* ZAP */
	DMU_OT_DSL_PERMS            /* ZAP */
	DMU_OT_ACL                  /* ACL */
	DMU_OT_SYSACL               /* SYSACL */
	DMU_OT_FUID                 /* FUID table (Packed NVLIST UINT8) */
	DMU_OT_FUID_SIZE            /* FUID table size UINT64 */
	DMU_OT_NEXT_CLONES          /* ZAP */
	DMU_OT_SCAN_QUEUE           /* ZAP */
	DMU_OT_USERGROUP_USED       /* ZAP */
	DMU_OT_USERGROUP_QUOTA      /* ZAP */
	DMU_OT_USERREFS             /* ZAP */
	DMU_OT_DDT_ZAP              /* ZAP */
	DMU_OT_DDT_STATS            /* ZAP */
	DMU_OT_SA                   /* System attr */
	DMU_OT_SA_MASTER_NODE       /* ZAP */
	DMU_OT_SA_ATTR_REGISTRATION /* ZAP */
	DMU_OT_SA_ATTR_LAYOUTS      /* ZAP */
	DMU_OT_SCAN_XLATE           /* ZAP */
	DMU_OT_DEDUP                /* fake dedup BP from ddt_bp_create() */
	DMU_OT_NUMTYPES
	//	/*
	//	 * Names for valid types declared with DMU_OT().
	//	 */
	DMU_OTN_UINT8_DATA      //  = DMU_OT(DMU_BSWAP_UINT8, B_FALSE),
	DMU_OTN_UINT8_METADATA  // = DMU_OT(DMU_BSWAP_UINT8, B_TRUE),
	DMU_OTN_UINT16_DATA     /// = DMU_OT(DMU_BSWAP_UINT16, B_FALSE),
	DMU_OTN_UINT16_METADATA // = DMU_OT(DMU_BSWAP_UINT16, B_TRUE),
	DMU_OTN_UINT32_DATA     //  = DMU_OT(DMU_BSWAP_UINT32, B_FALSE),
	DMU_OTN_UINT32_METADATA //  = DMU_OT(DMU_BSWAP_UINT32, B_TRUE),
	DMU_OTN_UINT64_DATA     //  = DMU_OT(DMU_BSWAP_UINT64, B_FALSE),
	DMU_OTN_UINT64_METADATA //  = DMU_OT(DMU_BSWAP_UINT64, B_TRUE),
	DMU_OTN_ZAP_DATA        //  = DMU_OT(DMU_BSWAP_ZAP, B_FALSE),
	DMU_OTN_ZAP_METADATA    // = DMU_OT(DMU_BSWAP_ZAP, B_TRUE)
//} dmu_object_type_t;
)

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
	Type               DmuObjectType      // dn type
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

const (
	CompressionInherit   = ZfsCompressionType(iota) // "ZIO_COMPRESS_INHERIT",
	ComperssionOn                                   // "ZIO_COMPRESS_ON"
	CompressionOff                                  // "ZIO_COMPRESS_OFF",
	CompressionLZJB                                 // "ZIO_COMPRESS_LZJB",
	CompressionEmpty                                // "ZIO_COMPRESS_EMPTY",
	CompressionGzip1                                // "ZIO_COMPRESS_GZIP_1",
	CompressionGzip2                                // "ZIO_COMPRESS_GZIP_2",
	CompressionGzip3                                // "ZIO_COMPRESS_GZIP_3",
	CompressionGzip4                                // "ZIO_COMPRESS_GZIP_4",
	CompressionGzip5                                // "ZIO_COMPRESS_GZIP_5",
	CompressionGzip6                                // "ZIO_COMPRESS_GZIP_6",
	CompressionGzip7                                // "ZIO_COMPRESS_GZIP_7",
	ComperssionGzip8                                // "ZIO_COMPRESS_GZIP_8",
	CompressionGzip9                                //"ZIO_COMPRESS_GZIP_9",
	CompressionLZE                                  // "ZIO_COMPRESS_ZLE",
	CompressionLZ4                                  // "ZIO_COMPRESS_LZ4",
	CompressionFunctions                            // "ZIO_COMPRESS_FUNCTIONS",
)

func (zct ZfsCompressionType) Decompress(dst []byte, src []byte) (int, error) {
	switch zct.String() {
	case "ZIO_COMPRESS_INHERIT":
		return 0, nil
	case "ZIO_COMPRESS_ON":
		return 0, nil
	case "ZIO_COMPRESS_LZ4":
		return func(dst []byte, src []byte) (int, error) {
			d, err := lz4.Decode(dst, src)
			if err != nil {
				return 0, err
			}
			return len(d), nil
		}(dst, src)
	case "ZIO_COMPRESS_ZLE":
		return 0, nil
	default:
		return func(dst []byte, src []byte) (int, error) { return copy(dst, src), nil }(dst, src)
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

func (bpp BlockPointerProps) Level() int {
	return int((uint64(bpp) &^ (1 << 63)) >> uint64(56))

}

func (bpp BlockPointerProps) Embedded() bool {
	return (uint64(bpp>>39) & 0x01) == 1
}

// Logical Size - size without compression (decompressed size)
func (bpp BlockPointerProps) Lsize() int {
	return int(uint8(bpp&0xff)+1) * 512
}

// Physical Size - size on disk
func (bpp BlockPointerProps) Psize() int {
	return int(uint8((bpp>>8)&0xff)+1) * 512
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

type ZfsVdev struct {
	Back io.ReadSeeker
	VDL  VdevLabel
}

func (z *ZfsVdev) Read() {
}
