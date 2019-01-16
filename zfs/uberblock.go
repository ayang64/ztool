package zfs

import (
	"fmt"
	"io"
	"time"
)

type ActiveUberBlock struct {
	AShift uint
	UberBlock
}

func (a *ActiveUberBlock) Lsize() int {
	return int(a.UberBlock.RootBP.Props.Lsize()) * (1 << a.AShift)
}

func (a *ActiveUberBlock) Psize() int {
	return int(a.UberBlock.RootBP.Props.Psize()) * (1 << a.AShift)
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
	SoftwareVersion  uint64       // FreeBSD is usually 5000
	PAD              [3]uint64    // padding
	CheckpointTx     uint64       // Checkpoint Transaction
}

func (ub *UberBlock) MOS(rs io.ReadSeeker) (*DnodePhys, error) {
	// the MOS is always element 1 in the UberBlocks's Vdev
	// list.
	return ub.RootBP.GetRootBlock(rs)
}

func (ub *UberBlock) String() string {
	return fmt.Sprintf("\nMagic: %08x (valid: %v), ", ub.Magic, ub.Magic == 0xbab10c) +
		fmt.Sprintf("Version: %d, ", ub.Version) +
		fmt.Sprintf("TrasnactionGroup: %d, ", ub.TransactionGroup) +
		fmt.Sprintf("Timestamp: %s (%d)\n", time.Unix(int64(ub.Timestamp), 0), ub.Timestamp) +
		fmt.Sprintf("GUID Sum: %x, ", ub.GuidSum) +
		fmt.Sprintf("SoftwareVersion: %d, ", ub.SoftwareVersion) +
		fmt.Sprintf("Checkpoint Transaction: %x\n", ub.CheckpointTx) +
		fmt.Sprintf("%s", ub.RootBP)
}
