package zfs

import (
	"encoding/binary"
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
	// ZFSDATA overrides hostname.
	if rc := os.Getenv("ZFSDATA"); rc != "" {
		return rc
	}
	if n, _ := os.Hostname(); n == "jade.ayan.net" {
		return "/obrovsky/recovery/zbackup0"
	}
	return "../zbackup0"
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
		{Name: "BlockPointer", Value: BlockPointer{}, ExpectedSize: 128},
		{Name: "DnodePhys", Value: DnodePhys{}, ExpectedSize: 512},
		{Name: "UberBlock", Value: UberBlock{}, ExpectedSize: 208},
		{Name: "VdevLabel", Value: VdevLabel{}, ExpectedSize: 262144},
		{Name: "VdevOffset", Value: VdevOffset{}, ExpectedSize: 16},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if typ := reflect.TypeOf(test.Value); typ.Size() != test.ExpectedSize {
				t.Fatalf("Size of %s is %d bytes, expected %d (%d bytes)",
					typ.Name(), typ.Size(), test.ExpectedSize, test.ExpectedSize/8)
				t.FailNow()
			}
		})
	}
}

func TestZfs(t *testing.T) {
	t.Logf("--------------------------------------")

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
		for idx, ub := range v.UberBlocks() {
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

	// read phy dnone from offsets.

	uberBlock, err := vdl[0].ActiveUberBlock()

	t.Logf("uberBlock: %s", uberBlock)

	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	t.Logf("ACTIVE UBER BLOCK: %#v", uberBlock)
	dn, err := uberBlock.MOS(r)

	if err != nil {
		t.Logf("uberBlock.MOS(r) failed: %v", err)
		t.FailNow()
	}

	t.Logf("### %#v", dn)
	t.Logf("compression type: %s", dn.Compress)
}
