package zfs

import (
	"encoding/binary"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestSizeofs(t *testing.T) {
	tests := []struct {
		Name         string
		Value        interface{}
		ExpectedSize uintptr
	}{
		{Name: "VdevOffset", Value: VdevOffset{}, ExpectedSize: 16},
		{Name: "BlockPointer", Value: BlockPointer{}, ExpectedSize: 128},
		{Name: "UberBlock", Value: UberBlock{}, ExpectedSize: 168},
		{Name: "VdevLabel", Value: VdevLabel{}, ExpectedSize: 262144},
		{Name: "DnodePhys", Value: DnodePhys{}, ExpectedSize: 512},
	}

	t.Logf("%d", 128<<10)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			typ := reflect.TypeOf(test.Value)
			t.Logf("sizeof(%s) = %d bytes aligned at %d", typ.Name(), typ.Size(), typ.Align())
			if typ.Size() != test.ExpectedSize {
				t.Fatalf("Size of %s is %d bytes, expected %d (%d bytes)", typ.Name(), typ.Size(), test.ExpectedSize, test.ExpectedSize/8)
				t.FailNow()
				return
			}
		})
	}
}

func TestZfs(t *testing.T) {

	testFilePath := func() string {
		if rc := os.Getenv("ZFSDATA"); rc != "" {
			return rc
		}
		return "../zbackup0"
	}

	r, err := os.Open(testFilePath())

	if err != nil {
		t.Fatalf("could not open vdev: %v", err)
		t.FailNow()
	}

	defer r.Close()

	vdl := [2]VdevLabel{}
	binary.Read(r, binary.LittleEndian, &vdl)

	for idx, v := range vdl {
		t.Logf("********** HEADER %d **********", idx)
		for idx, ub := range v.UberBlock() {
			if ub.Magic != 0xbab10c {
				continue
			}
			t.Logf("---------- UBER BLOCK %03d ----------", idx)
			t.Logf("ub.Magic = %010x (valid = %v)", ub.Magic, ub.Magic == 0xbab10c)
			t.Logf("ub.Version = %d", ub.Version)
			t.Logf("ub.RootBP = %#v", ub.RootBP)
			t.Logf("ub.RootBP.Padding = %v", ub.RootBP.Padding)
			t.Logf("ub.RootBP.FillCount = %d", ub.RootBP.FillCount)
			t.Logf("ub.RootBP.BirthTransactionGroup = %0d", ub.RootBP.BirthTransactionGroup)
			t.Logf("ub.RootBP.Props = %d (%b)", ub.RootBP.Props, ub.RootBP.Props)
			t.Logf("ub.RootBP.Props.Endian() = %s", ub.RootBP.Props.Endian())
			t.Logf("ub.RootBP.Props.Checksum() = %d", ub.RootBP.Props.Checksum())
			t.Logf("ub.RootBP.Props.Type() = %d", ub.RootBP.Props.Type())
			t.Logf("ub.RootBP.Props.Compression() = %d", ub.RootBP.Props.Compression())
			// t.Logf("ub.RootBP.Props.Compression = %q", ub.RootBP.Props.Compression)

			for idx, vd := range ub.RootBP.Vdevs {
				t.Logf("vdev %03d: %#v, gang = %v, disk offset = %d", idx, vd, vd.Gang(), vd.Block())
				t.Logf("    Asize: %d, Size: %d", vd.Asize(), vd.Size)
			}

			tim := time.Unix(int64(ub.Timestamp), 0)
			t.Logf("ub.RootBP.Timestamp = %s (%d)", tim, ub.Timestamp)
		}
	}

	// read phy dnone from offsets.

	uberBlock := vdl[0].UberBlock()
	for idx := range uberBlock[0].RootBP.Vdevs {
		offset := uberBlock[0].RootBP.Vdevs[idx].Block()
		t.Logf("%d", offset)

		if _, err := r.Seek(int64(offset), 0); err != nil {
			t.Logf("r.Seek(%d, 0) returned %v", offset, err)
			t.FailNow()
		}

		// read a DnodePhys into memory
		dn := DnodePhys{}
		binary.Read(r, binary.LittleEndian, &dn)

		t.Logf("%#v", dn)
	}
}
