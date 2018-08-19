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
		{Name: "VdevLabel", Value: VdevLabel{}, ExpectedSize: 168},
	}

	t.Logf("%d", 128<<10)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			typ := reflect.TypeOf(test.Value)
			t.Logf("sizeof(%s) = %d (%d bytes), aligned at %d", typ.Name(), typ.Size(), typ.Size()/8, typ.Align())
			if typ.Size() != test.ExpectedSize {
				t.Fatalf("Size of %s is %d (%d bytes), expected %d (%d bytes)", typ.Name(), typ.Size(), typ.Size()/8, test.ExpectedSize, test.ExpectedSize/8)
				t.FailNow()
				return
			}
		})
	}
}

func TestZfs(t *testing.T) {
	r, err := os.Open("../zbackup0")

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
			t.Logf("ub.RootBP.Props.Compression = %q", ub.RootBP.Props.Compression)
			t.Logf("ub.RootBP.Props = %#v", ub.RootBP.Props)
			t.Logf("ub.RootBP.BirthTransactionGroup = %d", ub.RootBP.BirthTransactionGroup)

			for idx, vd := range ub.RootBP.Vdevs {
				t.Logf("vdev %03d: %#v, gang = %v, disk offset = %d", idx, vd, vd.Gang(), vd.Block())
			}

			tim := time.Unix(int64(ub.RootBP.Birth), 0)
			t.Logf("ub.RootBP.Birth = %s (%d)", tim, ub.RootBP.Birth)
		}
	}
}
