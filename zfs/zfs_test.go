package zfs_test

import (
	"encoding/binary"
	"github.com/ayang64/ztool/zfs"
	_ "github.com/pierrec/lz4"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// This is the big test.
func TestZFS(t *testing.T) {
	tests := []struct {
		Name string
		Path string
	}{
		{Name: "simple 100mb file", Path: "../test-data/uncompressed-simple.image"},
		{Name: "original disk", Path: "../../zbackup-editable"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			// get a new filesystem.
			fs, err := zfs.New(zfs.WithPath(test.Path))

			if err != nil {
				t.Fatal(err)
			}

			ashift, _ := fs.AShift()

			t.Logf("ashift = %d", ashift)

			ubs, err := fs.UberBlocks()
			ubs = ubs
			if err != nil {
				t.Fatal(err)
			}

			ub, err := fs.ActiveUberBlock()

			if err != nil {
				t.Fatal(err)
			}

			// t.Logf("%#v", ub)
			ub = ub

			/*
				for _, ub := range ubs {
					t.Logf("%#v", ub)
					t.Logf(">> %v, %d, %d, compression: %s (%d)",
						ub.RootBP.Props.Embedded(),
						ub.RootBP.Props.Psize(),
						ub.RootBP.Props.Lsize(),
						ub.RootBP.Props.Compression(),
						ub.RootBP.Props.Compression())
				}
			*/
			for idx, vd := range ub.RootBP.Vdevs {
				t.Logf("vdev %03d (%d): %#v, gang = %v, disk offset = %d", vd.VDEV, idx, vd, vd.Gang(), vd.Block())
				t.Logf("    Asize: %d, Size: %d", vd.Asize(), vd.Size)
			}
			t.Logf("%s", ub.RootBP)

			t.Logf("ub.Psize() = %d\n", ub.Psize())
			t.Logf("ub.Lsize() = %d\n", ub.Lsize())
			t.Logf("ub.RootBP.Props.Psize() = %d", ub.RootBP.Props.Psize())
			t.Logf("ub.RootBP.Props.Lsize() = %d", ub.RootBP.Props.Lsize())

			t.Logf("%d/%d", ub.RootBP.Vdevs[0].Block(), ub.RootBP.Vdevs[0].Offset)
		})
	}
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

func TestFindUberBlocks(t *testing.T) {
	tests := []struct {
		Name string
		Path string
	}{
		{Name: "simple 100mb file", Path: "../test-data/uncompressed-simple.image"},
		{Name: "original disk", Path: "../../zbackup-editable"},
	}

	for _, test := range tests {

		t.Run(test.Name, func(t *testing.T) {
			r, err := os.Open(test.Path)

			if err != nil {
				t.Fatal(err)
			}

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

		})
	}
}
