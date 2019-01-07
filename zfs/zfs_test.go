package zfs_test

import (
	"encoding/binary"
	"github.com/ayang64/zfstool/zfs"
	"github.com/pierrec/lz4"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
func testFilePath() string {
	orStr := func(strs ...string) string {
		for _, s := range strs {
			if s != "" {
				return s
			}
		}

		return ""
	}
	return orStr(os.Getenv("ZFSDATA"), "/obrovsky/recovery/zbackup0")
}

func TestSeek(t *testing.T) {
	inf, err := os.Open(testFilePath())

	if err != nil {
		t.Fatalf("error: %v", err)
		t.FailNow()
	}

	s, err := inf.Stat()
	if err != nil {
		t.Fatalf("error: %v", err)
		t.FailNow()
	}

	t.Logf("s = %#v", s)
}

func TestSizeofs(t *testing.T) {
	tests := []struct {
		Name         string
		Value        interface{}
		ExpectedSize uintptr
	}{
		{Name: "BlockPointer", Value: zfs.BlockPointer{}, ExpectedSize: 128},
		{Name: "DnodePhys", Value: zfs.DnodePhys{}, ExpectedSize: 512},
		{Name: "UberBlock", Value: zfs.UberBlock{}, ExpectedSize: 208},
		{Name: "VdevLabel", Value: zfs.VdevLabel{}, ExpectedSize: 262144},
		{Name: "VdevOffset", Value: zfs.VdevOffset{}, ExpectedSize: 16},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if typ := reflect.TypeOf(test.Value); typ.Size() != test.ExpectedSize {
				t.Fatalf("Size of %s is %d bytes; expected %d",
					typ.Name(), typ.Size(), test.ExpectedSize)
				t.FailNow()
			}
		})
	}
}

func fileReader(t *testing.T) (*os.File, error) {
	return os.Open(testFilePath())
}

func TestZfs(t *testing.T) {
	t.Logf("--------------------------------------")
	rc, err := os.Open(testFilePath())
	if err != nil {
		t.Fatalf("could not open vdev: %v", err)
		t.FailNow()
	}
	rc = rc
}

func TestFindUberBlocks(t *testing.T) {

	r := func() *os.File {
		r, err := fileReader(t)

		if err != nil {
			t.Fatal(err)
			t.FailNow()
		}
		return r
	}()

	defer r.Close()

	vdl := [2]zfs.VdevLabel{}
	binary.Read(r, binary.LittleEndian, &vdl)

	for idx, v := range vdl {
		t.Logf("********** LABEL %d **********", idx)

		ubs := v.UberBlocks()

		t.Logf("THERE ARE %d UBER BLOCKS", len(ubs))
		for idx, ub := range ubs {
			t.Logf("---------- UBER BLOCK %03d ----------", idx)

			t.Logf("%s", &ub)

			for idx, vd := range ub.RootBP.Vdevs {
				t.Logf("vdev %03d: %#v, gang = %v, disk offset = %d", idx, vd, vd.Gang(), vd.Block())
				t.Logf("    Asize: %d, Size: %d", vd.Asize(), vd.Size)
			}
			tim := time.Unix(int64(ub.Timestamp), 0)
			t.Logf("ub.RootBP.Timestamp = %s (%d)", tim, ub.Timestamp)
		}
	}

}

func TestFindMOS(t *testing.T) {
	r := func() *os.File {
		r, err := fileReader(t)
		if err != nil {
			t.Fatal(err)
			t.FailNow()
		}
		return r
	}()

	vdl := [2]zfs.VdevLabel{}
	binary.Read(r, binary.LittleEndian, &vdl)

	ub := vdl[1].UberBlocks()

	t.Logf("%v", ub)

	uberBlock, err := vdl[0].ActiveUberBlock()

	t.Logf("uberBlock: %s", uberBlock)

	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	t.Logf("ACTIVE UBER BLOCK:\n%s", uberBlock)

	offset := uberBlock.RootBP.Vdevs[1].Block()

	kb := offset / 1024
	mb := kb / 1024
	gb := mb / 1024

	t.Logf("OFFSET: %d (%dmb) (%dgb)", offset, mb, gb)

	finfo, err := r.Stat()

	if err != nil {
		t.Logf("stat failed: %v", err)
	}

	t.Logf("file size: %d", finfo.Size())
	if finfo.Size() < int64(offset) {
		t.Logf("attempting to go past EOF.")
	}

	if _, err := r.Seek(int64(offset), 0); err != nil {
		t.Logf("r.Seek(%d, 0) returned %v", offset, err)
		t.FailNow()
	}

	// read a DnodePhys into memory
	dn := zfs.DnodePhys{}

	lzr := lz4.NewReader(r)
	if err := binary.Read(lzr, binary.LittleEndian, &dn); err != nil {
		t.Logf("binary.Read() failed: %v", err)
	}
	// binary.Read(r, binary.LittleEndian, &dn)

	t.Logf("------\n%#v", dn)

	t.Logf("shift: %d", dn.IndirectBlockShift)
	t.Logf("compression type: %s", dn.Compress)
}
