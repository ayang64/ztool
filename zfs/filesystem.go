package zfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/ayang64/ztool/zfs/internal/nvlist"
)

type cache struct {
	ashift *uint64
}

type Filesystem struct {
	rs     io.ReadSeeker
	nvlist nvlist.List

	vdl VdevLabel

	cache
}

func WithNVList(nvl map[string]interface{}) func(*Filesystem) error {
	return func(fs *Filesystem) error {
		fs.nvlist = nvl
		return nil
	}
}

func WithPath(path string) func(*Filesystem) error {
	return func(fs *Filesystem) error {
		fh, err := os.Open(path)

		if err != nil {
			return err
		}

		fs.rs = fh
		return nil
	}
}

func WithReadSeeker(rs io.ReadSeeker) func(*Filesystem) error {
	return func(fs *Filesystem) error {
		fs.rs = rs
		return nil
	}
}

func New(opts ...func(*Filesystem) error) (*Filesystem, error) {
	rc := Filesystem{}
	for _, opt := range opts {
		if err := opt(&rc); err != nil {
			return nil, err
		}
	}

	if err := rc.LoadVdevLabel(); err != nil {
		return nil, err
	}

	if err := rc.readNVP(); err != nil {
		return nil, err
	}

	return &rc, nil
}

func (fs *Filesystem) LoadVdevLabel() error {
	// get nvlist
	fs.rs.Seek(0, 0)
	if err := binary.Read(fs.rs, binary.LittleEndian, &fs.vdl); err != nil {
		return err
	}

	return nil
}

func (fs *Filesystem) ActiveUberBlock() (*ActiveUberBlock, error) {
	ubs, err := fs.UberBlocks()

	if err != nil {
		return nil, err
	}

	var txg uint64
	var idx int

	for i := range ubs {
		if ubs[i].TransactionGroup <= txg {
		}
		idx, txg = i, ubs[i].TransactionGroup
	}

	ashift, err := fs.AShift()

	if err != nil {
		return nil, err
	}

	aub := ActiveUberBlock{
		AShift:    ashift,
		UberBlock: ubs[idx],
	}

	return &aub, nil
}

func (fs *Filesystem) UberBlocks() ([]UberBlock, error) {
	// the ashift determines our block size.  this also determines how many
	// uberblocks we can fit in our uberblock array.
	ashift, err := fs.AShift()
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(fs.vdl.UberBlockBuf[:])

	rc := []UberBlock{}

	// the uber block section is 128 << 10 bytes in size.
	//
	// FIXME: this feels hamfisted.  i feel like there is a better way to express
	// this.
	for i, rsize, nrecords := 0, (1 << ashift), (128<<10)/(1<<ashift); i < nrecords; i++ {
		offset := int64(i * rsize)
		r.Seek(offset, 0)

		var ub UberBlock

		binary.Read(r, binary.LittleEndian, &ub)
		rc = append(rc, ub)
	}

	return rc, nil
}

func (fs *Filesystem) AShift() (uint, error) {
	if fs.cache.ashift != nil {
		return uint(*fs.cache.ashift), nil
	}

	ashift, found := fs.nvlist.Find("ashift")

	if !found {
		return 0, fmt.Errorf("ashift not found")
	}

	fs.cache.ashift = new(uint64)

	*fs.cache.ashift = ashift.(uint64)

	return uint(*fs.cache.ashift), nil
}

func (fs *Filesystem) readNVP() error {

	r := bytes.NewReader(fs.vdl.NVP[:])

	l, err := nvlist.Read(r)

	if err != nil {
		return err
	}

	fs.nvlist = l
	return nil
}
